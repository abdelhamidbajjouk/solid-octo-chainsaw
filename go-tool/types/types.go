package types

type JSON struct {
	ID    string     `json:"id"`
	Items []JSONItem `json:"items"`
}

type JSONItem struct {
	Name string `json:"name"`
	Src  string `json:"src"`
}
