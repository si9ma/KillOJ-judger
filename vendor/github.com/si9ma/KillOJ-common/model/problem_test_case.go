package model

type ProblemTestCase struct {
	ID             int    `gorm:"column:id;primary_key" json:"id"`
	ProblemID      int    `gorm:"column:problem_id" json:"problem_id"`
	InputData      string `gorm:"column:input_data" json:"input_data"`
	ExpectedOutput string `gorm:"column:expected_output" json:"expected_output"`
}

// TableName sets the insert table name for this struct type
func (p *ProblemTestCase) TableName() string {
	return "problem_test_case"
}
