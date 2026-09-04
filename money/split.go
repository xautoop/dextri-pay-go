package money

import (
	"fmt"
	"math/big"
	"sort"
)

const basisPointDenominator = uint32(10_000)

// BasisPointShare 描述一个收款目标及其万分比，所有份额之和必须为 10000。
type BasisPointShare struct {
	Key string
	BPS uint32
}

// AtomicShare 是可直接提交给 Pay 结算接口的原子金额分配结果。
type AtomicShare struct {
	Key          string
	AmountAtomic string
}

// SplitAtomic 以原子单位计算比例，并把每项结果限制到 outputDecimals 位小数。
// 除不尽的最小展示单位按余数从大到小分配；余数相同时保持入参顺序。
func SplitAtomic(totalAtomic string, assetDecimals, outputDecimals uint32, shares []BasisPointShare) ([]AtomicShare, error) {
	if outputDecimals > assetDecimals {
		return nil, fmt.Errorf("output decimals cannot exceed asset decimals")
	}
	total, ok := new(big.Int).SetString(totalAtomic, 10)
	if !ok || total.Sign() <= 0 {
		return nil, fmt.Errorf("total atomic amount must be a positive integer")
	}
	if len(shares) == 0 {
		return nil, fmt.Errorf("shares must not be empty")
	}
	bpsTotal := uint64(0)
	seen := make(map[string]struct{}, len(shares))
	for _, share := range shares {
		if share.Key == "" || share.BPS == 0 {
			return nil, fmt.Errorf("share key and positive BPS are required")
		}
		if _, exists := seen[share.Key]; exists {
			return nil, fmt.Errorf("share keys must be unique")
		}
		seen[share.Key] = struct{}{}
		bpsTotal += uint64(share.BPS)
	}
	if bpsTotal != uint64(basisPointDenominator) {
		return nil, fmt.Errorf("share BPS must sum to %d", basisPointDenominator)
	}

	quantum := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(assetDecimals-outputDecimals)), nil)
	totalUnits, remainder := new(big.Int).QuoRem(total, quantum, new(big.Int))
	if remainder.Sign() != 0 {
		return nil, fmt.Errorf("total amount exceeds requested output precision")
	}
	type calculated struct {
		index     int
		units     *big.Int
		remainder *big.Int
	}
	denominator := big.NewInt(int64(basisPointDenominator))
	calculatedShares := make([]calculated, len(shares))
	allocatedUnits := new(big.Int)
	for index, share := range shares {
		product := new(big.Int).Mul(totalUnits, big.NewInt(int64(share.BPS)))
		units, shareRemainder := new(big.Int).QuoRem(product, denominator, new(big.Int))
		calculatedShares[index] = calculated{index: index, units: units, remainder: shareRemainder}
		allocatedUnits.Add(allocatedUnits, units)
	}
	missingUnits := new(big.Int).Sub(totalUnits, allocatedUnits)
	order := append([]calculated(nil), calculatedShares...)
	sort.SliceStable(order, func(i, j int) bool { return order[i].remainder.Cmp(order[j].remainder) > 0 })
	for cursor := new(big.Int); cursor.Cmp(missingUnits) < 0; cursor.Add(cursor, big.NewInt(1)) {
		index := order[int(cursor.Int64())%len(order)].index
		calculatedShares[index].units.Add(calculatedShares[index].units, big.NewInt(1))
	}

	result := make([]AtomicShare, len(shares))
	for index, share := range shares {
		amount := new(big.Int).Mul(calculatedShares[index].units, quantum)
		result[index] = AtomicShare{Key: share.Key, AmountAtomic: amount.String()}
	}
	return result, nil
}
