package taxes

var EstIncomeRateList{{.Year}} = []*EstvIncomeRate{
{{range $eir := .List}}
	{
		CommuneID: {{$eir.CommuneID}},
		CommuneName: "{{$eir.CommuneName}}",
		CommuneMultiplier: {{$eir.CommuneMultiplier}},
	},{{end}}
}