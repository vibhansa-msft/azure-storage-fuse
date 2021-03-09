package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type Arith int

type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}

func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func (t *Arith) Divide(args *Args, quo *Quotient) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	return nil
}

type TestST struct {
	FieldA string `yaml:"A"`
	FieldB string `yaml:"B"`
	FieldC struct {
		FieldC1 int `yaml:"C.1"`
		FieldC2 int `yaml:"C.2"`
		FieldC3 struct {
			FieldC31 string `yaml:"C.3.1"`
		} `yaml:"C3"`
	} `yaml:"C"`
	FieldD []int `yaml:"D"`
}

func GetTestST(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Called")
	//*reply = Caller.RemoteAddr().String()

	t1 := &TestST{
		FieldA: "Hello",
		FieldB: "World",
	}
	t1.FieldC.FieldC1 = 10
	t1.FieldC.FieldC2 = 20
	t1.FieldC.FieldC3.FieldC31 = "Enough"
	t1.FieldD = append(t1.FieldD, 10)
	t1.FieldD = append(t1.FieldD, 20)

	/*d, err := yaml.Marshal(&t1)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Fprintf(w, string(d))
	*/
	//fmt.Println(string(d))
	json.NewEncoder(w).Encode(t1)

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println(GetIP(r))
	fmt.Println("Endpoint Hit: homePage")
}

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func main() {
	/*arith := new(Arith)
	rpc.Register(arith)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)*/

	//http.HandleFunc("/", rootHandler)
	http.HandleFunc("/GetTestST", GetTestST)
	log.Fatal(http.ListenAndServe(":1234", nil))
}
