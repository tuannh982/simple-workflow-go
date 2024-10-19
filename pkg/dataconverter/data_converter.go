package dataconverter

type DataConverter interface {
	Unmarshal(data []byte, v any) error
	Marshal(v any) ([]byte, error)
}
