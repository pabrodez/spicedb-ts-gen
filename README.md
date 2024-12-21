# spicedb-ts-gen 
Go package that parses a SpiceDb schema and outputs a string containing Typescript definitions.

Using the generated definitions requires the npm module `@authzed/authzed-node` 

Example using [this example schema](https://github.com/authzed/examples/blob/main/schemas/user-defined-roles/schema-and-data.yaml):
```go
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	schemaConverter "github.com/pabrodez/spicedb-ts-gen/converter"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: go run main.go <folder path> <schema file name>")
	}
	folderPath := os.Args[1]
	schemaFileName := os.Args[2]

	if strings.TrimSpace(folderPath) == "" {
		log.Fatal("Folder path is empty")
	}

	if strings.TrimSpace(schemaFileName) == "" {
		log.Fatal("Schema file name is empty")
	}

	fmt.Print(schemaConverter.GenerateDefinitionFromFS(os.DirFS(folderPath), schemaFileName))
}
```
```bash
go run main.go "." "schema.zed"
```

```typescript
import { v1 } from "@authzed/authzed-node";

type ResourcePermissionMap = {
  user: "",
  project: "create_issue" | "create_role",
  role: "delete" | "add_user" | "add_permission" | "remove_permission",
  issue: "assign" | "resolve" | "create_comment" | "project_comment_deleter",
  comment: "delete",
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
```

The helper class `PermissionRequest` can then be used like so:
```typescript
import { v1 } from '@authzed/authzed-node';

const client = v1.NewClient();
const { promises: promiseClient } = client;

const userAssignIssueRequest = new PermissionRequest<'user', 'issue'>()
    .from('user', 'user2')
    .to('issue', 'issue3')
    .withPermission('assign')
    .build();

const result = await promiseClient.checkPermission(userAssignIssueRequest);
```

Would like to: 
- Support Lookup requests 