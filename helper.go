package main

import(
	"fmt"
) 


func Printlnf(obj interface{}, message ...string) {
	if len(message) > 0 {
		fmt.Printf(message[0]+"\n", obj)
	} else {
		fmt.Printf("%T %+v\n", obj, obj)
	}
	
}
