package store

import (
	"fmt"

	"github.com/eveisesi/krinder"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Mongo Operators
const (
	equal            string = "$eq"
	greaterthan      string = "$gt"
	greaterthanequal string = "$gte"
	in               string = "$in"
	lessthan         string = "$lt"
	lessthanequal    string = "$lte"
	notequal         string = "$ne"
	notin            string = "$nin"
	and              string = "$and"
	or               string = "$or"
	exists           string = "$exists"
)

func BuildMongoFilters(operators ...*krinder.Operator) primitive.D {

	var ops = make(primitive.D, 0)
	for _, a := range operators {
		switch a.Operation {
		case krinder.EqualOp:
			ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: equal, Value: a.Value}}})
		case krinder.NotEqualOp:
			ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: notequal, Value: a.Value}}})
		case krinder.GreaterThanOp:
			ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: greaterthan, Value: a.Value}}})
		case krinder.GreaterThanEqualToOp:
			ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: greaterthanequal, Value: a.Value}}})
		case krinder.LessThanOp:
			ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: lessthan, Value: a.Value}}})
		case krinder.LessThanEqualToOp:
			ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: lessthanequal, Value: a.Value}}})
		case krinder.ExistsOp:
			ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: exists, Value: a.Value.(bool)}}})
		case krinder.OrOp:
			switch o := a.Value.(type) {
			case []*krinder.Operator:
				arr := make(primitive.A, 0)

				for _, op := range o {
					arr = append(arr, BuildMongoFilters(op))
				}

				ops = append(ops, primitive.E{Key: or, Value: arr})
			}

		case krinder.AndOp:
			switch o := a.Value.(type) {
			case []*krinder.Operator:
				arr := make(primitive.A, 0)
				for _, op := range o {
					arr = append(arr, BuildMongoFilters(op))
				}

				ops = append(ops, primitive.E{Key: and, Value: arr})
			}

		case krinder.InOp:
			switch o := a.Value.(type) {
			case []krinder.OpValue:
				arr := make(primitive.A, 0)
				for _, value := range o {
					arr = append(arr, value)
				}

				ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: in, Value: arr}}})
			default:
				panic(fmt.Sprintf("valid type %#T supplied, expected one of [[]krinder.OpValue]", o))
			}
		case krinder.NotInOp:
			switch o := a.Value.(type) {
			case []krinder.OpValue:
				arr := make(primitive.A, 0)
				for _, value := range o {
					arr = append(arr, value)
				}

				ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: notin, Value: arr}}})
			default:
				panic(fmt.Sprintf("valid type %#T supplied, expected one of [[]krinder.OpValue]", o))
			}
		}
	}

	return ops

}

func BuildMongoFindOptions(ops ...*krinder.Operator) *options.FindOptions {
	var opts = options.Find()
	for _, a := range ops {
		switch a.Operation {
		case krinder.LimitOp:
			opts.SetLimit(a.Value.(int64))
		case krinder.SkipOp:
			opts.SetSkip(a.Value.(int64))
		case krinder.OrderOp:
			opts.SetSort(primitive.D{primitive.E{Key: a.Column, Value: a.Value}})
		}
	}

	return opts
}
