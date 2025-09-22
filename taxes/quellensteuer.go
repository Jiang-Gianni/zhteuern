package taxes

type QuellenSteuer struct {
	Start              int
	PercentageTimes100 int
}

var QuellenSteuerListByYear = map[int][]*QuellenSteuer{
	2024: QuellenSteuerList24,
	2025: QuellenSteuerList25,
}

func QuellenSteuerPercentage(year int, salary float32) int {
	qsList, ok := QuellenSteuerListByYear[year]
	if !ok {
		return 0
	}
	if salary >= float32(qsList[len(qsList)-1].Start) {
		return qsList[len(qsList)-1].PercentageTimes100
	}
	i := 0
	start, end, mid := 0, len(qsList)-1, 0
	for {
		i++
		mid = start + (end-start)/2
		if float32(qsList[mid].Start) > salary {
			end = mid
			continue
		}
		if float32(qsList[mid+1].Start) > salary {
			return qsList[mid].PercentageTimes100
		}
		start = mid
	}
}
