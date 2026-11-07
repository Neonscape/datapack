package arraypreservation

import (
 "path/filepath"
 "testing"
 "cuelang.org/go/cue"
 "cuelang.org/go/cue/cuecontext"
 "cuelang.org/go/cue/load"
)

// These are real CUE files, not handcrafted codec/DTO examples. The external
// references are legal CUE yet rejected by this application's expression scope.
func TestArrayFixturesAreLegalCUE(t *testing.T) {
 for _, name := range []string{"arrays.cue", "fks.cue", "targets.cue", "rejected-external.cue"} {
  t.Run(name, func(t *testing.T) {
   file, err := filepath.Abs(name); if err != nil {t.Fatal(err)}
   instances := load.Instances([]string{file}, &load.Config{Dir: filepath.Dir(file)})
   if len(instances)!=1 {t.Fatalf("instances: %d",len(instances))}
   value := cuecontext.New().BuildInstance(instances[0])
   if err := value.Err(); err != nil {t.Fatal(err)}
   rows := value.LookupPath(cue.MakePath(cue.Str("rows")))
   if err := rows.Validate(cue.Concrete(true)); err != nil {t.Fatal(err)}
   data,err:=rows.MarshalJSON();if err!=nil {t.Fatal(err)}
   t.Logf("legal concrete CUE rows: %s",data)
  })
 }
}
