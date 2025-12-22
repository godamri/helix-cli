package ast

import (
	"fmt"
)

type Injector struct {
	FilePath string
}

func NewInjector(filePath string) *Injector {
	return &Injector{FilePath: filePath}
}

func (i *Injector) InjectEntityWiring(entityName, entityCamel string) error {
	fmt.Println("\nMANUAL WIRING REQUIRED")
	fmt.Println("---------------------------------------------------------")
	fmt.Printf("Open 'cmd/server/main.go' and add the following lines:\n\n")

	fmt.Printf("// Repository\n")
	fmt.Printf("repo%s := repository.New%sRepository(clientMain)\n\n", entityName, entityName)

	fmt.Printf("// Service\n")
	fmt.Printf("svc%s := service.New%sService(repo%s)\n\n", entityName, entityName, entityName)

	fmt.Printf("// Handler (HTTP)\n")
	fmt.Printf("httpHandler%s := handler.New%sHandler(svc%s)\n", entityName, entityName, entityName)
	fmt.Printf("r.Route(\"/v1/%ss\", func(r chi.Router) {\n", entityCamel)
	fmt.Printf("\tr.Post(\"/\", httpHandler%s.Create)\n", entityName)
	fmt.Printf("\tr.Get(\"/\", httpHandler%s.List)\n", entityName)
	fmt.Printf("\tr.Get(\"/{id}\", httpHandler%s.GetByID)\n", entityName)
	fmt.Printf("\tr.Put(\"/{id}\", httpHandler%s.Update)\n", entityName)
	fmt.Printf("\tr.Delete(\"/{id}\", httpHandler%s.Delete)\n", entityName)
	fmt.Printf("})\n")

	fmt.Println("---------------------------------------------------------")
	return nil
}
