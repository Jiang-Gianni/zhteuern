package taxes

type ContributionRates struct {
	AHV int
	ALV int
}

var (
	// 1st pillar and unemployment insurance are fixed by year
	ContributionRatesByYear = map[int]*ContributionRates{
		2024: ContributionRates2024,
		2025: ContributionRates2025,
	}
	ContributionRates2024 = &ContributionRates{
		AHV: 530,
		ALV: 110,
	}
	ContributionRates2025 = &ContributionRates{
		AHV: 530,
		ALV: 110,
	}
)
