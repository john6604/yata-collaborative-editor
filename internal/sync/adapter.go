package sync

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
)

type ConvertedOperation struct {
	Type   string
	Insert protocol.InsertOperation
	Delete protocol.DeleteOperation
}

func InsertOperationConverter(payload protocol.UpdatePayload) (ConvertedOperation, error) {

	var operation protocol.InsertOp

	err := json.Unmarshal(payload.Operation, &operation)

	if err != nil {
		return ConvertedOperation{}, err
	}

	content := []rune(operation.Character)

	if len(content) != 1 {
		return ConvertedOperation{}, errors.New("invalid_payload")
	}

	if operation.NewID.ClientID == "" || operation.OriginID.ClientID == "" || operation.RightID.ClientID == "" {
		return ConvertedOperation{}, errors.New("missing_field")
	}

	insertOp := protocol.NewInsertOperation(operation.NewID, operation.OriginID, operation.RightID, content[0])

	return ConvertedOperation{
		Type:   protocol.OpInsert,
		Insert: *insertOp,
		Delete: protocol.DeleteOperation{},
	}, nil
}

func DeleteOperationConverter(payload protocol.UpdatePayload) (ConvertedOperation, error) {

	var operation protocol.DeleteOp

	err := json.Unmarshal(payload.Operation, &operation)

	if err != nil {
		return ConvertedOperation{}, err
	}

	if operation.TargetID.ClientID == "" {
		return ConvertedOperation{}, errors.New("missing_field")
	}

	deleteOp := protocol.NewDeleteOperation(operation.TargetID)

	return ConvertedOperation{
		Type:   protocol.OpDelete,
		Insert: protocol.InsertOperation{},
		Delete: *deleteOp,
	}, nil
}

func ConvertUpdateOperation(payload protocol.UpdatePayload) (ConvertedOperation, error) {

	var operation protocol.OperationEnvelope
	var convertedOperation ConvertedOperation
	var errConversion error

	err := json.Unmarshal(payload.Operation, &operation)

	if err != nil {
		return ConvertedOperation{}, err
	}

	formattedType := strings.TrimSpace(operation.Type)

	if formattedType == "" {
		return ConvertedOperation{}, errors.New("missing_field")
	}

	switch formattedType {
	case protocol.OpInsert:
		convertedOperation, errConversion = InsertOperationConverter(payload)
		if errConversion != nil {
			return ConvertedOperation{}, errConversion
		}
	case protocol.OpDelete:
		convertedOperation, errConversion = DeleteOperationConverter(payload)
		if errConversion != nil {
			return ConvertedOperation{}, errConversion
		}
	case protocol.OpSync1:
		return ConvertedOperation{}, errors.New("unsupported_conversion")
	case protocol.OpSync2:
		return ConvertedOperation{}, errors.New("unsupported_conversion")
	default:
		return ConvertedOperation{}, errors.New("invalid_payload")
	}

	return convertedOperation, nil
}

func ApplyConvertedOperation(doc *document.Document, convertedOperation ConvertedOperation) error {

	switch convertedOperation.Type {
	case protocol.OpInsert:
		insert := convertedOperation.Insert
		err := doc.RemoteInsert(insert.NewID, insert.OriginID, insert.RightID, insert.Content)
		if err != nil {
			if err.Error() != "Pending value." {
				return err
			}
		}
	case protocol.OpDelete:
		delete := convertedOperation.Delete
		err := doc.RemoteDelete(delete.TargetID)
		if err != nil {
			return err
		}
	default:
		return errors.New("unsupported_operation")
	}

	return nil
}
