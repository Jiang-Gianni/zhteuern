package taxes

var QuellenSteuerList{{.Year}} = []*QuellenSteuer{
{{range $qs := .List}}
	{
		Start: {{$qs.Start}},
		PercentageTimes100: {{$qs.PercentageTimes100}},
	},{{end}}
}