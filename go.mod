module github.com/vinymeuh/radiogagad

go 1.14

require (
	github.com/vinymeuh/chardevgpio v0.0.0-20200401082432-55ce1062ebc8
	gopkg.in/yaml.v2 v2.2.8
	periph.io/x/periph v3.6.2+incompatible
)

//replace github.com/vinymeuh/chardevgpio => ../chardevgpio
