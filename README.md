# dianella

Smooth scripting inside Go for devops-ish automation work. A library of functions and types which simplify calling external processes and error handling.

Example:

```Go
func main() {

	flag.StringVar(&environment, "clubname", "", "one of: cricket, netball, frizbee")
	flag.Parse()

	begin("Fetch the club membership Information").
		set("member_ids", keysOfMap(memberIds)).
		set("env", "{{.Flags.environment}}").
		
		expand(`
            select * from club where member_id in (
            {{ range $index, $id := .Var.member_ids }}
                {{ if $index }},{{ end }} 
                '{{.}}'
            {{end}}      );
			`, "members.sql").
		bash("sqlite_wrapper.sh {{.Var.env}} members.sql").
		end()
}
```

TODO @Peter

# References

## Dianella longifolia

“Smooth Flax Lily”

![Dianella Longifolia](./dianella.jpg "")

LILY TO 1 METRE TALL

Long narrow, flax-like leaves forming tufts, bright blue flowers on branched stems to 1 metre high followed by attractive blue edible berries. Good rockery plant, and host to Yellow banded Dart butterfly. Leaves can be used to weave baskets.

[PPNN](https://ppnn.org.au/plantlist/dianella-longifolia/)
