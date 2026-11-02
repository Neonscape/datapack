package cueschema

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"datapack/internal/document"
	"datapack/internal/perftrace"
	"fmt"
	"path/filepath"
)

func Load(table document.Table, packageName string) (cue.Value, error) {
	defer perftrace.Start("cueschema.Load")()
	config := &load.Config{Dir: filepath.Dir(table.Path), Package: packageName}
	if source, ok := document.SchemaCapsule(table); ok {
		config.Overlay = map[string]load.Source{table.Path: load.FromBytes(source)}
	}
	instances := load.Instances([]string{"."}, config)
	if len(instances) != 1 {
		return cue.Value{}, fmt.Errorf("load table package: expected one package, found %d", len(instances))
	}
	value := cuecontext.New().BuildInstance(instances[0])
	if err := value.Err(); err != nil {
		return cue.Value{}, err
	}
	return value, nil
}
