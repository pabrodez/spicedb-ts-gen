package converter_test

import (
	"strings"
	"testing"
	"testing/fstest"

	schemaConverter "github.com/pabrodez/spicedb-ts-gen/converter"
)

const schema string = `
/**
 * user represents a user that can be granted role(s)
 */
definition user {}

definition org {
  relation member: user
  relation boss: user

  permission owner = boss
}

/**
 * document represents a document protected by Authzed.
 */
definition document {
    /**
     * writer indicates that the user is a writer on the document.
     */
    relation writer: user

    /**
     * reader indicates that the user is a reader on the document.
     */
    relation reader: user

    /**
     * edit indicates that the user has permission to edit the document.
     */
    permission edit = writer

    /**
     * view indicates that the user has permission to view the document, if they
     * are a reader *or* have edit permission.
     */
    permission view = reader + edit
}
`

const expectedTypescriptDefinitions string = `
import { v1 } from "@authzed/authzed-node";

type ResourcePermissionMap = {
  user: "",
  org: "owner",
  document: "edit" | "view",
}

type ResourceType = keyof ResourcePermissionMap;

export class PermissionRequest<S extends ResourceType, R extends ResourceType> {
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
};
`

func TestSpiceDbSchemaToTypescript(t *testing.T) {
	fs := fstest.MapFS{
		"schema.zed": &fstest.MapFile{
			Data: []byte(schema),
		},
	}
	got, err := schemaConverter.GenerateDefinitionFromFS(fs, "schema.zed")
	got = strings.TrimSpace(got)

	if err != nil {
		t.Fatal(err)
	}

	want := strings.TrimSpace(expectedTypescriptDefinitions)

	if got != want {
		t.Errorf("got: \n%s\n want: \n%s", got, want)
	}
}
