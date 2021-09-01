package mongodb

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/koderover/zadig/pkg/microservice/aslan/config"
	"github.com/koderover/zadig/pkg/microservice/aslan/core/common/repository/models"
	mongotool "github.com/koderover/zadig/pkg/tool/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type WorkLoadsStatColl struct {
	*mongo.Collection

	coll string
}

func NewWorkLoadsStatColl() *WorkLoadsStatColl {
	name := models.WorkLoadStat{}.TableName()
	return &WorkLoadsStatColl{Collection: mongotool.Database(config.MongoDatabase()).Collection(name), coll: name}
}

func (c *WorkLoadsStatColl) Create(args *models.WorkLoadStat) error {
	if args == nil {
		return errors.New("nil WorkLoadsCounter args")
	}
	_, err := c.InsertOne(context.TODO(), args)
	return err
}

func (c *WorkLoadsStatColl) Find(cluster string, namespace string) (*models.WorkLoadStat, error) {
	query := bson.M{}

	query["namespace"] = namespace

	if cluster != "" {
		query["cluster_id"] = cluster
	}

	resp := new(models.WorkLoadStat)

	err := c.FindOne(context.TODO(), query).Decode(resp)
	return resp, err
}

func (c *WorkLoadsStatColl) UpdateWorkloads(args *models.WorkLoadStat) error {
	query := bson.M{"namespace": args.Namespace, "cluster_id": args.ClusterID}
	change := bson.M{"$set": bson.M{
		"workloads": args.Workloads,
	}}
	_, err := c.UpdateOne(context.TODO(), query, change, options.Update().SetUpsert(true))
	return err
}
