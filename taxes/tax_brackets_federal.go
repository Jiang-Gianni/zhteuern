package taxes

type FederalRates struct {
	UpTo                 int
	AdditionalPercentage float32
	Base                 int
}

func GetBaseFederalSingle(year int, salary float32) (float32, *FederalRates) {
	fList, ok := FederalRatesSingleByYear[year]
	if !ok {
		return 0, &FederalRates{}
	}
	return getBaseFederal(fList, salary)
}

func GetBaseFederalMarriedChildren(year int, salary float32) (float32, *FederalRates) {
	fList, ok := FederalRatesMarriedChildrenByYear[year]
	if !ok {
		return 0, &FederalRates{}
	}
	return getBaseFederal(fList, salary)
}

func getBaseFederal(fList []*FederalRates, salary float32) (taxes float32, out *FederalRates) {
	for i, f := range fList {
		if i+1 < len(fList) && float32(fList[i+1].UpTo) > salary {
			taxes = float32(f.Base) + (salary-float32(f.UpTo))*f.AdditionalPercentage/100
			return taxes, f
		}
	}
	return 0, &FederalRates{}
}

var FederalRatesSingleByYear = map[int][]*FederalRates{
	2024: {
		{UpTo: 0, AdditionalPercentage: 0, Base: 0},
		{UpTo: 15000, AdditionalPercentage: 0.77, Base: 0},
		{UpTo: 32800, AdditionalPercentage: 0.88, Base: 137},
		{UpTo: 42900, AdditionalPercentage: 2.64, Base: 226},
		{UpTo: 57200, AdditionalPercentage: 2.97, Base: 603},
		{UpTo: 75200, AdditionalPercentage: 5.94, Base: 1138},
		{UpTo: 81000, AdditionalPercentage: 6.60, Base: 1483},
		{UpTo: 107400, AdditionalPercentage: 8.80, Base: 3225},
		{UpTo: 139600, AdditionalPercentage: 11.00, Base: 6059},
		{UpTo: 182600, AdditionalPercentage: 13.20, Base: 10789},
		{UpTo: 783200, AdditionalPercentage: 13.20, Base: 90068},
		{UpTo: 783300, AdditionalPercentage: 11.50, Base: 90080},
	},

	/*
		why is there a 100 CHF wide tax bracket 783200-783300?
		and why not just merge the previous 2 brackets with the same additional percentage (13.20)?
	*/

	2025: {
		{UpTo: 0, AdditionalPercentage: 0, Base: 0},
		{UpTo: 15200, AdditionalPercentage: 0.77, Base: 0},
		{UpTo: 33200, AdditionalPercentage: 0.88, Base: 139},
		{UpTo: 43500, AdditionalPercentage: 2.64, Base: 229},
		{UpTo: 58000, AdditionalPercentage: 2.97, Base: 612},
		{UpTo: 76100, AdditionalPercentage: 5.94, Base: 1150},
		{UpTo: 82000, AdditionalPercentage: 6.60, Base: 1500},
		{UpTo: 108800, AdditionalPercentage: 8.80, Base: 3269},
		{UpTo: 141500, AdditionalPercentage: 11.00, Base: 6146},
		{UpTo: 184900, AdditionalPercentage: 13.20, Base: 10920},
		{UpTo: 793300, AdditionalPercentage: 13.20, Base: 91229},
		{UpTo: 793400, AdditionalPercentage: 11.50, Base: 91241},
	},
}

var FederalRatesMarriedChildrenByYear = map[int][]*FederalRates{
	2024: {
		{UpTo: 0, AdditionalPercentage: 0, Base: 0},
		{UpTo: 29300, AdditionalPercentage: 1.00, Base: 0},
		{UpTo: 52700, AdditionalPercentage: 2.00, Base: 234},
		{UpTo: 60500, AdditionalPercentage: 3.00, Base: 390},
		{UpTo: 78100, AdditionalPercentage: 4.00, Base: 918},
		{UpTo: 93600, AdditionalPercentage: 5.00, Base: 1538},
		{UpTo: 107200, AdditionalPercentage: 6.00, Base: 2218},
		{UpTo: 119000, AdditionalPercentage: 7.00, Base: 2926},
		{UpTo: 128800, AdditionalPercentage: 8.00, Base: 3612},
		{UpTo: 136600, AdditionalPercentage: 9.00, Base: 4236},
		{UpTo: 142300, AdditionalPercentage: 10.00, Base: 4749},
		{UpTo: 146300, AdditionalPercentage: 11.00, Base: 5149},
		{UpTo: 148300, AdditionalPercentage: 12.00, Base: 5369},
		{UpTo: 150300, AdditionalPercentage: 13.00, Base: 5609},
		{UpTo: 928600, AdditionalPercentage: 13.00, Base: 106788},
		{UpTo: 928700, AdditionalPercentage: 11.50, Base: 106801},
	},

	2025: {
		{UpTo: 0, AdditionalPercentage: 0, Base: 0},
		{UpTo: 29700, AdditionalPercentage: 1.00, Base: 0},
		{UpTo: 53400, AdditionalPercentage: 2.00, Base: 237},
		{UpTo: 61300, AdditionalPercentage: 3.00, Base: 395},
		{UpTo: 79100, AdditionalPercentage: 4.00, Base: 929},
		{UpTo: 94900, AdditionalPercentage: 5.00, Base: 1561},
		{UpTo: 108600, AdditionalPercentage: 6.00, Base: 2246},
		{UpTo: 120500, AdditionalPercentage: 7.00, Base: 2960},
		{UpTo: 130500, AdditionalPercentage: 8.00, Base: 3660},
		{UpTo: 138300, AdditionalPercentage: 9.00, Base: 4284},
		{UpTo: 144200, AdditionalPercentage: 10.00, Base: 4815},
		{UpTo: 148200, AdditionalPercentage: 11.00, Base: 5215},
		{UpTo: 150300, AdditionalPercentage: 12.00, Base: 5446},
		{UpTo: 152300, AdditionalPercentage: 13.00, Base: 5686},
		{UpTo: 940800, AdditionalPercentage: 13.00, Base: 108191},
		{UpTo: 940900, AdditionalPercentage: 11.50, Base: 108204},
	},
}
