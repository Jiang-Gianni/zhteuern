package main

// (Cantonal, Federal) limits by year

var deductionLimitByYear = map[int]map[string][2]float32{
	2024: deductionLimit2024,
	2025: deductionLimit2025,
}

var deductionLimit2024 = map[string][2]float32{
	deductionTransport:       {5200, 3200},
	deductionMeal:            {3200, 3200},
	deductionProfession:      {4000, 4000},
	deductionThirdPillar:     {7056, 7056},
	deductionHealthInsurance: {2900, 1800},
}

var deductionLimit2025 = map[string][2]float32{
	deductionTransport:       {5200, 3300},
	deductionMeal:            {3200, 3200},
	deductionProfession:      {4000, 4000},
	deductionThirdPillar:     {7258, 7258},
	deductionHealthInsurance: {2900, 1800},
}
