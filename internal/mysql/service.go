package mysql

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eveisesi/krinder"
)

func BuildFilters(s sq.SelectBuilder, operators ...*krinder.Operator) sq.SelectBuilder {
	for _, a := range operators {
		if !a.Operation.IsValid() {
			continue
		}

		switch a.Operation {
		case krinder.EqualOp:
			s = s.Where(sq.Eq{a.Column: a.Value})
		case krinder.NotEqualOp:
			s = s.Where(sq.NotEq{a.Column: a.Value})
		case krinder.GreaterThanEqualToOp:
			s = s.Where(sq.GtOrEq{a.Column: a.Value})
		case krinder.GreaterThanOp:
			s = s.Where(sq.Gt{a.Column: a.Value})
		case krinder.LessThanEqualToOp:
			s = s.Where(sq.LtOrEq{a.Column: a.Value})
		case krinder.LessThanOp:
			s = s.Where(sq.Lt{a.Column: a.Value})
		case krinder.InOp:
			s = s.Where(sq.Eq{a.Column: a.Value.(interface{})})
		case krinder.NotInOp:
			s = s.Where(sq.NotEq{a.Column: a.Value.([]interface{})})
		case krinder.LikeOp:
			s = s.Where(sq.Like{a.Column: fmt.Sprintf("%%%v%%", a.Value)})
		case krinder.OrderOp:
			s = s.OrderBy(fmt.Sprintf("%s %s", a.Column, a.Value))
		case krinder.LimitOp:
			s = s.Limit(uint64(a.Value.(int64)))
		case krinder.SkipOp:
			s = s.Offset(uint64(a.Value.(int64)))
		}
	}

	return s

}
