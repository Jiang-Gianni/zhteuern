-- name: InsertTS :exec
insert into tax_simulation(tax_simulation_id) values( ? );

-- name: GetTS :one
select
    tax_simulation_id,
    version,
    gross_salary,
    ahv_beitrag,
	alv_beitrag,
    ktg_beitrag,
    bvg_beitrag,
    taxable_salary,
    year,
	commune_id,
	investment,
    deduction_other,
    deduction_transport,
	deduction_profession,
    deduction_third_pillar,
    deduction_health_insurance,
    deduction_meal
from tax_simulation where tax_simulation_id = ?;

-- name: UpdateTSDeduction :exec
update tax_simulation
set
    deduction_other = ?,
    deduction_transport = ?,
	deduction_profession = ?,
    deduction_third_pillar = ?,
    deduction_health_insurance = ?,
    deduction_meal = ?,
    version = version + 1
where tax_simulation_id = ?;

-- name: UpdateTSIncome :exec
update tax_simulation
set
    gross_salary = ?,
    ktg_beitrag = ?,
    bvg_beitrag = ?,
    taxable_salary = ?,
    year = ?,
    commune_id = ?,
    version = version + 1
where tax_simulation_id = ?;

-- name: UpdateTSInvestment :exec
update tax_simulation
set
    investment = ?,
    version = version + 1
where tax_simulation_id = ?;