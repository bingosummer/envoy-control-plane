package utils

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

func MustMarshalAny(m proto.Message) *any.Any {
	pbst, err := ptypes.MarshalAny(m)
	if err != nil {
		panic(err)
	}

	return pbst
}
