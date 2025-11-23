package attendance

// DataJSON là struct cấp cao nhất cho RPC request
type DataJSON struct {
	ID      int    `json:"id"`
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  Params `json:"params"`
}

// Params chứa các tham số chi tiết cho method "call"
type Params struct {
	// Args là một mảng chứa:
	// 1. Một mảng các ID nhân viên (ví dụ: [6303]) -> []int
	// 2. Một chuỗi chỉ định hành động/method cụ thể -> string
	// Vì nó là một mảng hỗn hợp, chúng ta sử dụng []interface{}
	Args   []interface{} `json:"args"`
	Model  string        `json:"model"`
	Method string        `json:"method"`
	Kwargs Kwargs        `json:"kwargs"`
}

// Kwargs chứa các tham số từ khóa
type Kwargs struct {
	Context Context `json:"context"`
}

// Context chứa thông tin ngữ cảnh môi trường và người dùng
type Context struct {
	Lang              string  `json:"lang"`
	TZ                string  `json:"tz"`
	UID               int     `json:"uid"`
	AllowedCompanyIDs []int   `json:"allowed_company_ids"`
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	EnLocationID      string  `json:"en_location_id"`
}
