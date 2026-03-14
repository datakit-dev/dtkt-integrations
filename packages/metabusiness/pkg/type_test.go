package pkg_test

import (
	"fmt"
	"testing"

	"github.com/datakit-dev/dtkt-integrations/metabusiness/pkg"
)

func Test_Types(t *testing.T) {
	reqTypes, err := pkg.LoadRequestTypes()
	if err != nil {
		t.Fatal(err)
	}

	for _, reqType := range reqTypes {
		if len(reqType.Roots) > 0 {
			fmt.Println(reqType.Name, reqType.Roots)
		}

		// fmt.Println(reqType.Name)
		// for name, typ := range reqType.Fields {
		// 	fmt.Println("\t", name, typ)
		// }

		// if len(reqType.Edges) > 0 {
		// 	fmt.Println("Edges:")
		// 	for _, edge := range reqType.Edges {
		// 		fmt.Println("\t", edge.Method, edge.Return)
		// 	}
		// }
	}
}
