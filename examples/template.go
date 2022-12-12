package main

import (
	"flag"
	. "github.com/birchb1024/dianella"
	"log"
	"os/user"
)

func main() {
	userid, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}
	var cricket bool
	flag.BoolVar(&cricket, "is_it_cricket", true, "default true")
	flag.Parse()

	s := BEGIN("variables and template example").
		Set("userid", userid).
		Bash("test $(uname) = Darwin && dscl . readall /users | grep -B 5 {{.Var.userid.Username}} | grep HomePhoneNumber").
		Expand("Is it Cricket? {{.Flag.is_it_cricket}} Args: {{ .Arg }}\n", "/dev/stdout").
		END()
	s = s
}
