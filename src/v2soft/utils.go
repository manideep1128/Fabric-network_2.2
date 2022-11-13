package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/chaincode/shim/ext/entities"
)

/**
	* @Description
	*	- Using the encryption key and IV creates the encryption entity.
	*
	* @params
	*	- []byte: transientMap
	*		- transientMap containing the encryption key and IV.
	*
	* @returns
	*	- payload: ent, nil
	*		- returns ent, nil containing the encryption entity for encryption and decryption.
	*
	* @returns
	*	- payload: nil, err
	*		- returns nil, err if there is error in creating the encryption entity.
	*			- Missing key: "Expected ENCKEY"
	*			- Missing IV: "Expected IV"
*/

func Encrypter(stub shim.ChaincodeStubInterface, bccspInst bccsp.BCCSP, transientMap map[string][]byte) (entities.EncrypterEntity, error) {
	transientMap, err := stub.GetTransient()
	if _, in := transientMap["ENCKEY"]; !in {
		return nil, errors.New("Expected ENCKEY")
	}

	if _, in := transientMap["IV"]; !in {
		return nil, errors.New("Expected IV")
	}

	// create the encrypter entity - we give it an ID, the bccsp instance, the key and (optionally) the IV
	ent, err := entities.NewAES256EncrypterEntity("ID", bccspInst, transientMap["ENCKEY"], transientMap["IV"])

	if err != nil {
		return nil, err
	}

	return ent, nil
}

/**
	* @Description
	*	- Using the encryption entity to encrypt and put the data in the blockchain.
	*
	* @params
	*	- []byte: transientMap
	*		- transientMap containing the encryption key and IV.
	*	- string: key
	*		- id against which the data is to be stored.
	*	- []byte: data
	*		- data which is to be stored in the blockchain.
	*
	* @returns
	*	- payload: { response returned from the blockchain }
	*
	* @returns
	*	- payload: nil, err
	*		- returns nil, err if there is error in creating the encryption entity or 
	*			encrypting the data.
	*			- Encryption entity generation failed: "Failed to generate ent"
	*			- Encryption failed: "Failed to create ciphertext"
*/

func encryptAndPutState(stub shim.ChaincodeStubInterface, bccspInst bccsp.BCCSP, transientMap map[string][]byte, key string, data []byte) error {
	ent, err := Encrypter(stub, bccspInst, transientMap)
	fmt.Sprintf("ent in put state function ::%s",ent)
	if err != nil {
		return errors.New("Failed to generate ent")
	}

	// Encrypting the data to be stored in the blockchain
	ciphertext, err := ent.Encrypt(data)
    fmt.Sprintf("Could not encrypt n put state, err %s", err)

	if err != nil {
		return errors.New("Failed to create ciphertext")
	}
	// Putting the encrypted data in the blockchain
	return stub.PutState(key, []byte(ciphertext))
}

/**
	* @Description
	*	- Using the encryption entity to fetch the data from the blockchain and decrypt.
	*
	* @params
	*	- []byte: transientMap
	*		- transientMap containing the encryption key and IV.
	*	- string: key
	*		- id for which the data is to be fetched.
	*
	* @returns
	*	- payload: []byte
	*		- return byte array containing the decrypted data from the blockchain for the
	*			specified key
	*
	* @returns
	*	- payload: nil, err
	*		- returns nil, err if there is error in creating the encryption entity or 
	*			no value against the key.
	*			- Encryption entity generation failed: "Failed to generate ent"
	*			- Record fetch failed: "Failed to get state for {key}"
	*			- No value against the key: "Nil value for {key}"
*/

func getStateAndDecrypt(stub shim.ChaincodeStubInterface, bccspInst bccsp.BCCSP, transientMap map[string][]byte, key string) ([]byte, error) {
	ent, err := Encrypter(stub, bccspInst, transientMap)
	if err != nil {
		return nil, errors.New("Failed to generate ent")
	}
	// Fetching the data from the blockchain against that specific key
	ciphertext, err := stub.GetState(key)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	// Check for null value against that specific key
	if len(ciphertext) == 0 {
		jsonResp := "{\"Error\":\"Nil value for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	// Decrypt and return the encrypted data received from the blockchain
	return ent.Decrypt(ciphertext)
}

func GetHistoryForKeyAndDecrypt(stub shim.ChaincodeStubInterface, bccspInst bccsp.BCCSP, transientMap map[string][]byte, key string) ([]byte, error) {
	ent, err := Encrypter(stub, bccspInst, transientMap)
	if err != nil {
		return nil, errors.New("Failed to generate ent")
	}

	// Decrypt and return the encrypted data received from the blockchain
	return ent.Decrypt([]byte(key))
}