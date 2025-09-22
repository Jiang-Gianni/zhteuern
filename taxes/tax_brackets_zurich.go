package taxes

type ZurichRates struct {
	ForTheNextCHF        int
	AdditionalPercentage int
}

func GetBaseZurichSingle(year int, salary float32) (float32, *ZurichRates) {
	zrList, ok := ZurichRatesSingleByYear[year]
	if !ok {
		return 0, &ZurichRates{}
	}
	return getBaseZurich(zrList, salary)
}

func GetBaseZurichMarriedChildren(year int, salary float32) (float32, *ZurichRates) {
	zrList, ok := ZurichRatesMarriedChildrenByYear[year]
	if !ok {
		return 0, &ZurichRates{}
	}
	return getBaseZurich(zrList, salary)
}

func getBaseZurich(zrList []*ZurichRates, salary float32) (taxes float32, out *ZurichRates) {
	for _, zr := range zrList {
		if salary <= float32(zr.ForTheNextCHF) {
			taxes += salary * float32(zr.AdditionalPercentage) / 100
			return taxes, zr
		}
		taxes += float32(zr.ForTheNextCHF*zr.AdditionalPercentage) / 100
		// subtract the amount already taxed at this current bracket
		salary -= float32(zr.ForTheNextCHF)
	}
	return 0, &ZurichRates{}
}

// 2024 and 2025 have the same values
var ZurichRatesSingle = []*ZurichRates{
	{ForTheNextCHF: 6900, AdditionalPercentage: 0},
	{ForTheNextCHF: 4900, AdditionalPercentage: 2},
	{ForTheNextCHF: 4800, AdditionalPercentage: 3},
	{ForTheNextCHF: 7900, AdditionalPercentage: 4},
	{ForTheNextCHF: 9600, AdditionalPercentage: 5},
	{ForTheNextCHF: 11000, AdditionalPercentage: 6},
	{ForTheNextCHF: 12900, AdditionalPercentage: 7},
	{ForTheNextCHF: 17400, AdditionalPercentage: 8},
	{ForTheNextCHF: 33600, AdditionalPercentage: 9},
	{ForTheNextCHF: 33200, AdditionalPercentage: 10},
	{ForTheNextCHF: 52700, AdditionalPercentage: 11},
	{ForTheNextCHF: 68400, AdditionalPercentage: 12},
	{ForTheNextCHF: 99999999, AdditionalPercentage: 13},
}

var ZurichRatesSingleByYear = map[int][]*ZurichRates{
	2024: ZurichRatesSingle,
	2025: ZurichRatesSingle,
}

var ZurichRatesMarriedChildren = []*ZurichRates{
	{ForTheNextCHF: 13900, AdditionalPercentage: 0},
	{ForTheNextCHF: 6300, AdditionalPercentage: 2},
	{ForTheNextCHF: 8000, AdditionalPercentage: 3},
	{ForTheNextCHF: 9700, AdditionalPercentage: 4},
	{ForTheNextCHF: 11100, AdditionalPercentage: 5},
	{ForTheNextCHF: 14300, AdditionalPercentage: 6},
	{ForTheNextCHF: 31800, AdditionalPercentage: 7},
	{ForTheNextCHF: 31900, AdditionalPercentage: 8},
	{ForTheNextCHF: 47900, AdditionalPercentage: 9},
	{ForTheNextCHF: 57200, AdditionalPercentage: 10},
	{ForTheNextCHF: 62100, AdditionalPercentage: 11},
	{ForTheNextCHF: 71600, AdditionalPercentage: 12},
	{ForTheNextCHF: 99999999, AdditionalPercentage: 13},
}

var ZurichRatesMarriedChildrenByYear = map[int][]*ZurichRates{
	2024: ZurichRatesMarriedChildren,
	2025: ZurichRatesMarriedChildren,
}
