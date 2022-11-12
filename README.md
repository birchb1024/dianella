# NAME

dianella - Smooth shell-like scripting inside Go. For devops-ish automation work.

A library of functions and types which simplify calling external processes and error handling.

## SYNOPSIS

```
import (
	"flag"
	. "github.com/birchb1024/dianella"
)

func main() {
	flag.Parse()
	s := BEGIN("example").
		Bash("ls -l").
		END()
	s = s
}
```

## DESCRIPTION

to do

### Types

#### Stepper

#### Step

### Built-in Methods

also to do

#### Before()
#### After()

#### BEGIN()
#### AND()
#### END()

#### Bash()
#### Sbash()
#### Call()

#### Expand()

### Extending Dianella 

with composition and overriding pure virtual functions

TODO 

Example:

```Go
type myStep struct {
	Step
	timestamp time.Time
	details   any
}

func MyBEGIN(desc string) *myStep {
	m := myStep{}
	m.Init(&m, desc)
	return &m
}
func (m *myStep) After() {
	fmt.Printf("%#v %s\n", m.details, time.Now().Sub(m.timestamp))
}
func (m *myStep) Before(info ...any) { m.timestamp = time.Now(); m.details = info }


func main() {
	var s Stepper
	s = MyBEGIN("Start example1").
		AND("Set a variable to the current date").
		Set("date", time.Now().String())

```

### EXAMPLES:

```Go
func main() {
	s := BEGIN("Start example1").
		Set("trace", true).
		Set("date", time.Now().String()).
		Call(func(s Stepper) Stepper {
			fmt.Printf("%#v\n", s.GetVar())
			return s
		}).
		Bash("date").
		Bash("echo {{.Var.date}}")
	tmpFile, s := s.Sbash("mktemp")
	tmpFile = strings.TrimSpace(tmpFile)
	s.Set("tmpFile", tmpFile).
		Expand("tmpFile - Date: {{.Var.date}}\n", tmpFile).
		Bash("cat {{.Var.tmpFile}}").
		Bash("rm -f {{.Var.tmpFile}}")
	s.END()
	s = s
}
```

## SEE ALSO

[script]()

## REFERENCES

### Dianella longifolia

“Smooth Flax Lily”

![Dianella Longifolia](./dianella.jpg "")

LILY TO 1 METRE TALL

Long narrow, flax-like leaves forming tufts, bright blue flowers on branched stems to 1 metre high followed by attractive blue edible berries. Good rockery plant, and host to Yellow banded Dart butterfly. Leaves can be used to weave baskets.

[PPNN](https://ppnn.org.au/plantlist/dianella-longifolia/)
