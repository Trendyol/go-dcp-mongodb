package mongodb

import "go.mongodb.org/mongo-driver/bson"

type OperationType string

const (
	Insert OperationType = "insert"
	Update OperationType = "update"
	Delete OperationType = "delete"
	Upsert OperationType = "upsert"
)

type Model interface {
	Convert() *ExecArgs
}

type Raw struct {
	ID        string
	Document  bson.M
	Operation OperationType
}

type ExecArgs struct {
	Collection string
	Document   bson.M
	Operation  OperationType
}

func (r *Raw) Convert() *ExecArgs {
	return &ExecArgs{
		Document:  r.Document,
		Operation: r.Operation,
	}
}
