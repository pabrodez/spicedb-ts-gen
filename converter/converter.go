package converter

import (
	"io/fs"
	"log"
	"strings"

	"github.com/authzed/spicedb/pkg/development"
	"github.com/authzed/spicedb/pkg/namespace"
	corev1 "github.com/authzed/spicedb/pkg/proto/core/v1"
	implv1 "github.com/authzed/spicedb/pkg/proto/impl/v1"
)

const (
	importAuthzedNodeTemplate     = "import { v1 } from \"@authzed/authzed-node\";\n"
	resourcePermissionMapTemplate = "type ResourcePermissionMap = {"
	resourceTypeTemplate          = "type ResourceType = keyof ResourcePermissionMap;\n"
	permissionRequestTemplate     = `export class PermissionRequest<S extends ResourceType, R extends ResourceType> {
  private subject?: v1.SubjectReference;
  private resource?: v1.ObjectReference;
  private permission?: ResourcePermissionMap[R];

  from(type: S, id: string): PermissionRequest<S, R> {
    this.subject = {
      object: {
        objectType: type,
        objectId: id
      }
    } as v1.SubjectReference;

    return this;
  }

  to(type: R, id: string): PermissionRequest<S, R> {
    this.resource = {
      objectType: type,
      objectId: id
    };
    return this;
  }

  withPermission(permission: ResourcePermissionMap[R]): PermissionRequest<S, R> {
    this.permission = permission;
    return this;
  }

  build(): v1.CheckPermissionRequest {
    if (!this.subject || !this.resource || !this.permission) {
      throw new Error('Incomplete permission request');
    }
    return v1.CheckPermissionRequest.create({
      resource: this.resource,
      permission: this.permission,
      subject: this.subject,
    })
  }
};`
)

func GenerateDefinitionFromFS(fileSystem fs.FS, schemaFile string) (string, error) {
	schemaStr, readErr := readSchema(fileSystem, schemaFile)

	if readErr != nil {
		return "", readErr
	}

	compiledSchema, devError, err := development.CompileSchema(schemaStr)

	if err != nil || devError != nil {
		log.Fatal("error compiling schema: %w \n %w", devError, err)
	}

	var resourcePermissionMap strings.Builder

	resourcePermissionMap.WriteString(resourcePermissionMapTemplate)

	for _, def := range compiledSchema.ObjectDefinitions {
		appendPermissions(&resourcePermissionMap, def.Name, def.Relation)
	}

	resourcePermissionMap.WriteString("\n}\n")

	return formatOutput(resourcePermissionMap.String()), nil
}

func appendPermissions(resourcePermissionMap *strings.Builder, resourceName string, relations []*corev1.Relation) {
	resourcePermissionMap.WriteString("\n  " + resourceName + ": ")
	permissionsOfKind := filterRelationsByKind(relations, implv1.RelationMetadata_PERMISSION)
	if len(permissionsOfKind) > 0 {
		for reli, rel := range permissionsOfKind {
			resourcePermissionMap.WriteString("\"" + rel.Name + "\"")

			if reli < len(permissionsOfKind)-1 {
				resourcePermissionMap.WriteString(" | ")
			}
		}
		resourcePermissionMap.WriteString(",")
	} else {
		resourcePermissionMap.WriteString("\"\",")
	}
}

func filterRelationsByKind(relations []*corev1.Relation, kind implv1.RelationMetadata_RelationKind) []*corev1.Relation {
	var filtered []*corev1.Relation
	for _, rel := range relations {
		if namespace.GetRelationKind(rel) == kind {
			filtered = append(filtered, rel)
		}
	}
	return filtered
}

func formatOutput(resourcePermissionMap string) string {
	return importAuthzedNodeTemplate +
		"\n" +
		resourcePermissionMap +
		"\n" +
		resourceTypeTemplate +
		"\n" +
		permissionRequestTemplate
}

func readSchema(fileSystem fs.FS, fileName string) (string, error) {
	readFile, readErr := fs.ReadFile(fileSystem, fileName)

	if readErr != nil {
		return "", readErr
	}

	return string(readFile), nil
}
