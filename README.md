# NAME

dianella - Smooth shell-like scripting inside Go. A library of functions and types which 
simplify calling external processes and error handling.

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

`dianella` provides facilities found in UNIX shell scripting languages within Go programs. Shell scripts interpreters
execute arbitrary programs in sequence. These 'commands' return an exit status to indicate if they encountered an error
condition during processing. The shell interpreter monitors the return status and, with appropriate option set, the 
shell script is terminated and an error is produced. Thus, the programmer is not required to check the status 
of each command. 

```shell
set -e          # tells the shell to stop on error
ls -l           # this line runs
ls /Bzzzz       # this line fails
date            # this line never runs
```

Idiomatic Go programs always require explicit checking of calls to functions whic typically return an error value

```Go
value, err : = someFunction(args)
if err != nil {
    // handle the error
}
```

`dianella` programs are structured like a shell:

```Go
import . "github.com/birchb1024/dianella"

var Stepper s = Step.BEGIN("").
    Bash("ls -l").
    Bash("ls /Bzzz").
    Bash("date").
    END()
```

The module provides a set of methods which take a `Stepper` receiver and return a `Stepper`. This allows sequences of
function calls to be chained together. The `BEGIN()` constructs a struct (type Step) and returns it. The `Step`
struct holds the result status of a step, error information and a description field. `Step` also holds a symbol 
table for variables which can be set by the programmer, and then used in subsequent steps.
```Go
	userid, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}

	s := BEGIN("variables and template example").
		Set("userid", userid).
		Bash("dscl . readall /users | grep -B 5 {{.Var.userid.Username}} | grep HomePhoneNumber").
		END()
```
In this example the variable `userid` contains a struct the `Set()` function assigns the struct to the `userid`
variable. Later, in the `Bash()` step, the template module injects the userid with `{{.Var.userid}}`. Refer to the 
`template` module [documentation](https://pkg.go.dev/text/template) for details.

Most shell script languages suffer from a dearth of usable functions and syntax
for manipulating data, be they strings or structures. `dianella` allows access to all the Go universe of proper
language facilities and modules, whilst writing what are, in essence, 'shell 'scripts'.

The `Step` struct also includes 

* `.Flag`  - a map with command-line options from from the `flag` module
* `.Arg`    - a slice of commad-line options from `flag.Args()`

These can be used :
```Go
	var cricket bool
	flag.BoolVar(&cricket, "is_it_cricket", true, "default true")
	flag.Parse()

	s := BEGIN("variables and template example").
		Bash("echo {{.Flag.is_it_cricket}} {{index .Arg 1}}").
		END()
```

### Types

#### `Step`

```
type Step struct {
	Arg         []string
	Flag        map[string]any
	Var         map[string]any
	description string
	err         error
	Self        Stepper
	status      int
}
```
This struct is the 'base class' for dianalla, it holds the basic information used in an execution. As functions are called,
information is added to the struct. The `Self` variable holds an interface object holding pointer
to the struct itself. This is initialised in the `BEGIN()` and `Init()` functions. 

#### `Stepper` interface

The `Stepper` interface provides an abstract data type to step structs. It allows the methods of struct `Step` to be pure
virtual. The code inside the Step methods makes calls to other functions via the `Self` variable. eg

```Go
func (s *Step) Sexpand(template string) string {
	s.Self.Before("Sexpand", template[:intMin(len(template)-1, 20)])
	defer s.Self.After()
	return Expando(template, s)
}
```
Here you can see that the virtual functions `Before()` and `After()` are called via `Self`. 

### Extending `dianella`

Since all the methods for `Step` are defined in the `Stepper` interface they can be overriden in a subtype.
The best way to eplain this is by an example. Adding new data to the `Step` struct is straightforwrd:
```Go
type myStep struct {
	Step
	timestamp time.Time
	details   any
	dbUrl     string
}
```
Because `myStep` include `Step` all the `Stepper` methods are available. To construct a new `myStep`, initialise with `Init()`
```Go
func MyBEGIN(desc, url string) *myStep {
	m := myStep{dbUrl: url}
	m.Init(&m, desc)
	return &m
}
```
Now we can override `Stepper` methods, let's measure execution times by overriding the Before() and After()
methods which are called by all methods
```Go
func (m *myStep) Before(info ...any) { 
	m.timestamp = time.Now()
    m.details = info 
}
func (m *myStep) After() {
	fmt.Printf("%#v %s\n", m.details, time.Now().Sub(m.timestamp))
}
```
This transforms the behaviour of `dianella` - the logging output is replaced with timing data. We could have 
both, by calling the parental type's methods from the subtype method e.g. `m.Step.Before(info)`.

We can also add new methods in the Stepper style by adding the to our subtype:

```Go
func (m *myStep) PostgreSQL(query string) Stepper {
	if m.Self.IsFailed() {
		return m
	}
	m.Self.Before("PosgreSQL", query)
	defer m.Self.After()

	data := [][]string{{"Name", "Runs"}, {"Hales", "7"}, {"Butler", "54"}}
	fmt.Printf("%#v", data)
	return m
}
```
The statement `if m.Self.IsFailed() { return m }` is essential in all step methods, because is a prior method has failed, 
this prevents execution of this method. Execution continues till the `END()` function is called where the program is
terminated.

Calling the new functions requires a `myStep` receiver because they are not in the `Stepper` interface:

```Go
func main() {
    flag.Parse()
    var s *myStep
    s = MyBEGIN("Start example1", "postgres://localhost:5234/mydatabase").
    AND("Database query")
    s.PostgreSQL("select * from batters").
    END()
```

### Built-in `Stepper` Methods


#### `Before()`
This method is called by all the other methods at the beginning of execution. The (*Step) method
prints log information if the `"trace"` variable is `true`. Doing a `Set("trace", false)` disables the tracing.

#### `After()`
This method is called by all the other methods at the end of execution.

#### `BEGIN()`
Returns a pointer to a new `Step` object.

#### `AND()`
Updates the step description.

#### `END()`
Finishes the execution, if there has been a failure in the previous step functions, an error message is printed, 
and the process terminates.

#### `Bash()`
Calls `/bin/bash` as a subprocess, passing the argument to `-c ` after it has been expanded by the template module 

#### `Sbash()`
Like `Bash()` but returns the stdout of the sub-process as a string. 

#### `Call()`
Calls a user-supplied function passing it the step. 

#### `Expand()`
This function expands a template via the Go template module and writes the outcome to a file. This is used for 
code or text generation for example: 
```Go
	s.AND("Fetch player info").
		Expand(`
			select * from players where player_id in ( {{ range $index, $id := .Var.playerIds }}
				{{ if $index }},{{ end }} {{/* the first item is 0 which is also false in the if */}}
				'{{.}}'
			{{end}}      );
        `, "players.sql")
```

#### `Sexpand()`
This is the same as `Expand()` but the output is returned in a string.

#### `IsFailed()`
Returns `true` if the step has an error or has non-zero status.

### EXAMPLES:

Refer to the [examples/](examples) directory in this repo for more examples.

A basic example:

```Go
func printAllVariables(s Stepper) Stepper {
	fmt.Printf("%#v\n", s.GetVar())
	return s
}
func main() {
	flag.Parse()
	var s Stepper = BEGIN("Start example1").
		// Set("trace", false).
		Set("date", time.Now().String()).
		Call(printAllVariables).
		Bash("date").
		Bash("echo {{.Var.date}}")
	tmpFile, s := s.Sbash("mktemp")
	tmpFile = strings.TrimSpace(tmpFile)
	s.Set("tmpFile", tmpFile).
		Expand("tmpFile - Date: {{.Var.date}}\n", tmpFile).
		Bash("cat {{.Var.tmpFile}}").
		Bash("rm -f {{.Var.tmpFile}}").
		END()
	s = s
}
```

## SEE ALSO

[bitfield script](github.com/bitfield/script)

## REFERENCES

### Dianella longifolia

“Smooth Flax Lily”

![Dianella Longifolia](./dianella.jpg "")

LILY TO 1 METRE TALL

Long narrow, flax-like leaves forming tufts, bright blue flowers on branched stems to 1 metre high followed by attractive blue edible berries. Good rockery plant, and host to Yellow banded Dart butterfly. Leaves can be used to weave baskets.

[PPNN](https://ppnn.org.au/plantlist/dianella-longifolia/)
