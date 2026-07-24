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

func EncodeInsertOperation(insertOp protocol.InsertOperation) ([]byte, error) {

	content := string(insertOp.Content)

	insertOperation := protocol.InsertOp{
		Type:      protocol.OpInsert,
		NewID:     insertOp.NewID,
		OriginID:  insertOp.OriginID,
		RightID:   insertOp.RightID,
		Character: content,
	}

	insertEnvelope, errEnvelope := json.Marshal(insertOperation)

	if errEnvelope != nil {
		return nil, errEnvelope
	}

	updatePayload := protocol.UpdatePayload{
		Operation: insertEnvelope,
	}

	updateEnvelope, errUpdateEnv := json.Marshal(updatePayload)

	if errUpdateEnv != nil {
		return nil, errUpdateEnv
	}

	envelope := protocol.Envelope{
		Version:     protocol.SupportedVersion,
		MessageType: protocol.TypeUpdate,
		Payload:     updateEnvelope,
	}

	envelopeBytes, errBytes := json.Marshal(envelope)

	if errBytes != nil {
		return nil, errBytes
	}

	return envelopeBytes, nil
}

func EncodeDeleteOperation(deleteOp protocol.DeleteOperation) ([]byte, error) {

	deleteOperation := protocol.DeleteOp{
		Type:     protocol.OpDelete,
		TargetID: deleteOp.TargetID,
	}

	deleteEnvelope, errEnvelope := json.Marshal(deleteOperation)

	if errEnvelope != nil {
		return nil, errEnvelope
	}

	updatePayload := protocol.UpdatePayload{
		Operation: deleteEnvelope,
	}

	updateEnvelope, errUpdateEnv := json.Marshal(updatePayload)

	if errUpdateEnv != nil {
		return nil, errUpdateEnv
	}

	envelope := protocol.Envelope{
		Version:     protocol.SupportedVersion,
		MessageType: protocol.TypeUpdate,
		Payload:     updateEnvelope,
	}

	envelopeBytes, errBytes := json.Marshal(envelope)

	if errBytes != nil {
		return nil, errBytes
	}

	return envelopeBytes, nil
}

func EncodeUpdateOperation(convertedOperation ConvertedOperation) ([]byte, error) {

	var operation []byte
	var errConversion error

	switch convertedOperation.Type {
	case protocol.OpInsert:
		operation, errConversion = EncodeInsertOperation(convertedOperation.Insert)
		if errConversion != nil {
			return nil, errConversion
		}
	case protocol.OpDelete:
		operation, errConversion = EncodeDeleteOperation(convertedOperation.Delete)
		if errConversion != nil {
			return nil, errConversion
		}
	default:
		return nil, errors.New("unsupported_operation")
	}

	return operation, nil
}
