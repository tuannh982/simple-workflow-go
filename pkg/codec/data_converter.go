package codec

type Codec interface {
	Unmarshal(data []byte, v any) error
	Marshal(v any) ([]byte, error)
}
