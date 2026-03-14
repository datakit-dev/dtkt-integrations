package v1beta1_test

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/linkedin/goavro/v2"
// )

// func Test_TableRead(t *testing.T) {

// }

// func Test_flattenAvroData(t *testing.T) {
// 	codec, err := goavro.NewCodec(`
// 	{
// 		"type": "record",
// 		"name": "LongList",
// 		"fields" : [
// 			{"name": "next", "type": ["null", "LongList"], "default": null}
// 		]
// 	}`)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	// Convert native Go form to binary Avro data
// 	binary := []byte{0x2, 0x2, 0x0}

// 	native, _, err := codec.NativeFromBinary(binary)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	t.Log(codec.TypeName())
// 	t.Log(native.(map[string]any))
// 	t.Log(flattenAvroRow(native.(map[string]any)))

// 	// t.Log(flattenAvroRow(map[string]any{
// 	// 	"foo": map[string]any{
// 	// 		"string": "bar",
// 	// 	},
// 	// 	"bar": map[string]any{
// 	// 		"record": map[string]any{
// 	// 			"baz": "bam",
// 	// 		},
// 	// 	},
// 	// }))
// }
