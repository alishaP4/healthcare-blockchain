package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
        "strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type THcareChaincode struct {
}

const initID = 0

const (
	EMRIdKey = "EMRIDKEY"
)

const (
	UserPrefix	= "USER_"
	EMRPrefix	= "EMR_"
)

// EMR struct
type EMR struct{
	ID   	int 	//EMR ID
	Owner	string
	Adder 	string
	Type 	int   	//1 for medical record & 2 for test report
	Content string 	//内容
	Date 	string 	//日期
	AuthorityList map[string]int 	//授权阅读的人员列表
}

type EMRList struct{
        EMRmap  map[int]EMR
}

// User struct
type User struct {
	Name	string
	Type	string  // patient, hospital, doctor, goverment and thirdparty 用户、医院、医生、政府和第三方检测机构
	EMRNum  int
//	ownEMR  map[string]int	// emr list for this person
}

// for test

func (t *THcareChaincode) getUserEMRNum(stub shim.ChaincodeStubInterface, name string) (int, error) {

	user, err := t.getUserByName (stub, name)
	if err != nil {
		return -1, errors.New("User not found")
	}
	i := user.EMRNum
	return i,nil
}

//func (t *THcareChaincode) getUserEMRList(stub shim.ChaincodeStubInterface, name string) (map[string]int, error) {
//
//	user, err := t.getUserByName (stub, name)
//	if err != nil {
//		return make(map[string]int), errors.New(" not found")
//	}
//
//	return user.ownEMR, nil

//}


func (t *THcareChaincode) setUserByName(stub shim.ChaincodeStubInterface, user User) error {
	key := UserPrefix + user.Name
	userBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("Error marshalling user in the setUserByName function: %s", err)
	}
	err = stub.PutState(key, userBytes)
	if err != nil {
		fmt.Errorf(err.Error())
		return err
	}
	fmt.Printf("store user:%s sucessfully", user.Name)
	return nil
}

func (t *THcareChaincode) getUserByName(stub shim.ChaincodeStubInterface, name string) (User, error) {
	key := UserPrefix + name
	Valbytes, err := stub.GetState(key)
	if err != nil {
		return User{}, fmt.Errorf("Failed to get state %s", key)
	}
	if Valbytes == nil {
		return User{}, errors.New("User2 not found")
	}
	var user User
	err = json.Unmarshal(Valbytes, &user)
	if err != nil {
		return user, fmt.Errorf("Error unmarshalling user: %s", err)
	}

	return user, nil
}

func (t *THcareChaincode) addUser(stub shim.ChaincodeStubInterface, type0 string, name0 string) error {

	user := User{Name: name0, Type: type0, EMRNum: 0};

   //     user.ownEMR = make(map[string]int)
   //     user.ownEMR["init"] = -1

	err := t.setUserByName(stub, user)
	if err != nil {
		return err
	}
	return nil
}

func (t *THcareChaincode) setEMRByID(stub shim.ChaincodeStubInterface, emr EMR) error {
	key := EMRPrefix + strconv.Itoa(emr.ID)
	emrBytes, err := json.Marshal(emr)
	if err != nil {
		return fmt.Errorf("Error marshalling EMR in the setEMRByID function: %s", err)
	}
	err = stub.PutState(key, emrBytes)
	if err != nil {
		fmt.Errorf(err.Error())
		return err
	}
	fmt.Printf("store EMR:%s sucessfully", emr.Owner)
	return nil
}


func (t *THcareChaincode) getEMRByID(stub shim.ChaincodeStubInterface, serID int) (EMR, error) {
	key := EMRPrefix + strconv.Itoa(serID)
	Valbytes, err := stub.GetState(key)
	if err != nil {
		return EMR{}, fmt.Errorf("Failed to get state %s", key)
	}
	if Valbytes == nil {
		return EMR{}, errors.New("EMR not found")
	}
	var emr EMR
	err = json.Unmarshal(Valbytes, &emr)
	if err != nil {
		return emr, fmt.Errorf("Error unmarshalling emr: %s", err)
	}

	return emr, nil
}

// 外部搜索函数,搜索单条EMR
func (t *THcareChaincode) searchEMRByID(stub shim.ChaincodeStubInterface, serID int, queryName string) (EMR, error) {
	key := EMRPrefix + strconv.Itoa(serID)
	Valbytes, err := stub.GetState(key)
	if err != nil {
		return EMR{}, fmt.Errorf("Failed to get state %s", key)
	}
	if Valbytes == nil {
		return EMR{}, errors.New("EMR2 not found")
	}
	var emr EMR
	err = json.Unmarshal(Valbytes, &emr)
	
	if err==nil {
		_, ok := emr.AuthorityList[queryName]
		if ok{
			return emr, nil
		}

		return EMR{}, errors.New("No Authority to Read this EMR")
	}

	return emr, fmt.Errorf("Error unmarshalling emr: %s", err)
	
}

// 外部搜索函数，获取某人的所有EMR
func (t *THcareChaincode) searchAllEMR(stub shim.ChaincodeStubInterface, ownName string, queryName string) (EMRList, error) {
	emrList := make(map[int]EMR)

	var user User
	user, err := t.getUserByName(stub, ownName)
        if err != nil {
                return EMRList{}, err
        }
	if user.EMRNum<0 {
		return EMRList{}, errors.New("This patient have no EMR")
	}
	
	//遍历EMR
	id := 1

        currentID, err := t.getCurrentRecordID(stub, EMRIdKey)
        if err != nil {
                return EMRList{}, err
        }

	for {
		if id > int(currentID) {
			break;
		}
		temEMR, err := t.getEMRByID(stub, id)
                if err != nil {
                        return EMRList{}, err
                }

		if strings.EqualFold(temEMR.Owner, ownName) {
			_, ok := temEMR.AuthorityList[queryName]
			if ok{
			// 当前EMR记录有授权
			emrList[id] = temEMR
			}
		}
                id++
	}

	switch len(emrList){
	case 0:
		return EMRList{}, errors.New("No authorized EMR for this patient")
	default:
                emrmap := EMRList{EMRmap: emrList}
		return emrmap, nil
	}

}


func (t *THcareChaincode) addEMR(stub shim.ChaincodeStubInterface, owner string, adder string, type0 int, content string, date string) error {

	curID, err := t.getNextRecordID(stub, EMRIdKey)

        if err != nil {
            return err
        }

	emr := EMR{ID: int(curID), Owner: owner, Adder: adder, Type: type0, Content: content, Date: date};

	emr.AuthorityList = make(map[string]int)

        emr.AuthorityList[adder] = 1
	emr.AuthorityList[owner] = 1

	err = t.setEMRByID(stub, emr)
	if err != nil {
		return err
	}

	// add emr to the user

	var user User
	user, err = t.getUserByName(stub, owner)

//	user.ownEMR[string(curID)] = int(curID)

	user.EMRNum = user.EMRNum + 1

	err = t.setUserByName(stub, user)
	if err != nil {
		return err
	}

	return nil
}

// 为某条记录添加阅读权限
func (t *THcareChaincode) addReadAuthority(stub shim.ChaincodeStubInterface, emrID int, toAuthorName string, queryName string) error {

	var emr EMR
	emr, err := t.getEMRByID(stub, emrID);

	if err!=nil {
		return errors.New("No authorized EMR for this patient")
	}

	if emr.Owner == queryName {
		_, ok := emr.AuthorityList[toAuthorName]
		if ok {
		}else{
			emr.AuthorityList[toAuthorName] = 1
			err = t.setEMRByID(stub, emr)
			if err != nil {
			return err
	}
	
		}
		return nil
	}

	return nil
}

// 为某个病人的所有记录添加权限
func (t *THcareChaincode) addAllReadAuthority(stub shim.ChaincodeStubInterface, toAuthorName string, queryName string) error {

	//遍历EMR
	id := 1

        currentID, err := t.getCurrentRecordID(stub, EMRIdKey)
        if err != nil {
                return err
        }

	for {
		if id > int(currentID) {
			break;
		}
		temEMR, err := t.getEMRByID(stub, id)
                if err != nil {
                        return err
                }

		if strings.EqualFold(temEMR.Owner, queryName) {
			temEMR.AuthorityList[toAuthorName] = 1 
		}
                id++

                err = t.setEMRByID(stub, temEMR)
                if err != nil {
                        return err
                }
	}

	
	return nil
}

// 为某条记录删除阅读权限
func (t *THcareChaincode) delReadAuthority(stub shim.ChaincodeStubInterface, emrID int, toAuthorName string, queryName string) error {

	var emr EMR
	emr, err := t.getEMRByID(stub, emrID);

	if strings.EqualFold(emr.Owner, queryName){
		_, ok := emr.AuthorityList[toAuthorName]
		if ok{
			delete(emr.AuthorityList, toAuthorName)
		}
	}

	err = t.setEMRByID(stub, emr)
	if err != nil {
		return err
	}
	return nil
}

// 删除对某个病人所有记录的权限
func (t *THcareChaincode) delAllReadAuthority(stub shim.ChaincodeStubInterface, toAuthorName string, queryName string) error {
	//遍历EMR
	id := 1

        currentID, err := t.getCurrentRecordID(stub, EMRIdKey)
        if err != nil {
                return err
        }

	for {
		if id > int(currentID) {
			break;
		}
		temEMR, err := t.getEMRByID(stub, id)
                if err != nil {
                        return err
                }


		if strings.EqualFold(temEMR.Owner, queryName) {
			_, ok := temEMR.AuthorityList[toAuthorName]
			if ok{
				delete(temEMR.AuthorityList, toAuthorName)
			}
		}
                id++
                
                err = t.setEMRByID(stub, temEMR)
                if err != nil {
                        return err
                }
	}

	
	return nil
}

// Init EMR ID
func (t *THcareChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	err := initializeRecordID(stub)
	if err != nil {
		return nil, err
	}
	return nil, nil
}


// ID
func initializeRecordID(stub shim.ChaincodeStubInterface) error {
	var ID uint64
	ID = initID
	IDStr := strconv.FormatUint(ID, 10)
	err := stub.PutState(EMRIdKey, []byte(IDStr))
	if err != nil {
		return fmt.Errorf("Cannot put EMRID because " + err.Error())
	}
	return nil
}

func (t *THcareChaincode) getNextRecordID(stub shim.ChaincodeStubInterface, key string) (uint64, error) {
	var ID uint64
	currentIDBytes, err := stub.GetState(key)
	if err != nil {
		return initID, fmt.Errorf("Cannot get next ID because " + err.Error())
	}

	if currentIDBytes == nil {
		return initID, fmt.Errorf("Cannot get current ID because " + err.Error())
	} else {
		id, err := strconv.ParseUint(string(currentIDBytes), 10, 64)
		if err != nil {
			return initID, fmt.Errorf("Cannot parse the current ID because " + err.Error())
		}
		id++
		ID = id
	}

	IDStr := strconv.FormatUint(ID, 10)
	err = stub.PutState(key, []byte(IDStr))
	if err != nil {
		return initID, fmt.Errorf("Cannot store the ID because " + err.Error())
	}
	return ID, nil
}


func (t *THcareChaincode) getCurrentRecordID(stub shim.ChaincodeStubInterface, key string) (uint64, error) {
	var ID uint64
	currentIDBytes, err := stub.GetState(key)
	if err != nil {
		return initID, fmt.Errorf("Cannot get next ID because " + err.Error())
	}

	if currentIDBytes == nil {
		return initID, fmt.Errorf("Cannot get current ID because " + err.Error())
	} else {
		id, err := strconv.ParseUint(string(currentIDBytes), 10, 64)
		if err != nil {
			return initID, fmt.Errorf("Cannot parse the current ID because " + err.Error())
		}
		ID = id
	}

	return ID, nil
}

// Invoke
func (t *THcareChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	switch function {
	case "addUser":
		if len(args) != 2 {
			return nil, errors.New("Incorrect number of arguments in addUser: expect 2")
		}
		type0 := args[0]
		name0 := args[1]
		err := t.addUser(stub, type0, name0)
		if err != nil {
			fmt.Println("addUser error: ", err)
		}
		return nil, err
	case "addEMR":
		if len(args) != 5 {
			return nil, errors.New("Incorrect number of arguments in addEMR: expect 5")
		}
		owner := args[0]
		adder := args[1]
		type0, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Println("addEMR error: ", err)
		}
		content := args[3]
		date := args[4]
		err = t.addEMR(stub, owner, adder, type0, content, date)
		if err != nil {
			fmt.Println("addEMR error: ", err)
		}
		return nil, err
	case "addSingleReadAuthority":
		if len(args) != 3 {
			return nil, errors.New("Incorrect number of arguments in addSingleReadAuthority: expect 3")
		}
		emrID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("addSingleReadAuthority error: ", err)
		}
		toAuthorName := args[1]
		queryName := args[2]
		err = t.addReadAuthority(stub, emrID, toAuthorName, queryName);
		if err != nil {
			fmt.Println("addSingleReadAuthority error: ", err)
		}
		return nil, err
	case "addAllReadAuthority":
		if len(args) != 2 {
			return nil, errors.New("Incorrect number of arguments in addAllReadAuthority: expect 2")
		}
		toAuthorName := args[0]
		queryName := args[1]
		err := t.addAllReadAuthority(stub, toAuthorName, queryName);
		if err != nil {
			fmt.Println("addAllReadAuthority error: ", err)
		}
		return nil, err
		
	case "delSingleReadAuthority":
		if len(args) != 3 {
			return nil, errors.New("Incorrect number of arguments in delSingleReadAuthority: expect 3")
		}
		emrID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("delSingleReadAuthority error: ", err)
		}
		toAuthorName := args[1]
		queryName := args[2]
		err = t.delReadAuthority(stub, emrID, toAuthorName, queryName);
		if err != nil {
			fmt.Println("delSingleReadAuthority error: ", err)
		}
		return nil, err
	case "delAllReadAuthority":
		if len(args) != 2 {
			return nil, errors.New("Incorrect number of arguments in delAllReadAuthority: expect 2")
		}
		toAuthorName := args[0]
		queryName := args[1]
		err := t.delAllReadAuthority(stub, toAuthorName, queryName);
		if err != nil {
			fmt.Println("delAllReadAuthority error: ", err)
		}
		return nil, err
	default:
		errMsg := "No such method in Invoke method: " + function
		fmt.Errorf(errMsg)
		return nil, errors.New(errMsg)
	}
	return nil, nil
}


// Query
func (t *THcareChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	switch function {
	case "getUserEMRNum":
		if len(args) != 1 {
			return nil,
			errors.New("Incorrect number of arguments. Expecting user's name to query")
		}
		userName := args[0]
		ownEMRNum, err := t.getUserEMRNum(stub, userName)
		if err != nil {
			return nil, err
		}
        response, err := json.Marshal(ownEMRNum)
		return response, err

	case "getAllEMR":
		if len(args) != 2 {
			return nil,
			errors.New("Incorrect number of arguments. Expecting patient's name and your name to query")
		}
		ownName := args[0]
		queryName:= args[1]
		emrList, err := t.searchAllEMR(stub, ownName, queryName)
		if err != nil {
			return nil, err
		}
                emrlist_test := []EMR{}
                for _,emr := range emrList.EMRmap {
                        emrlist_test = append(emrlist_test, emr)
                }
                response, err := json.Marshal(emrlist_test)
		return response, err
		
	case "getSingleEMR":
		if len(args) != 2 {
			return nil,
			errors.New("Incorrect number of arguments. Expecting EMRID and your name to query")
		}
		emrID, err := strconv.Atoi(args[0])
		if err != nil {
			return nil, err
		}
		queryName := args[1]
		sp, err := t.searchEMRByID(stub, emrID, queryName)
		if err != nil {
			return nil, err
		}

		response, err := json.Marshal(sp)
		fmt.Printf("Query Response:%s\n", response)
		return response, err
	default:
		errMsg := "No such method in Query method:" + function
		fmt.Errorf(errMsg)
		return nil, errors.New(errMsg)
	}
	return nil, nil
}



func main() {
	err := shim.Start(new(THcareChaincode))
	if err != nil {
		fmt.Printf("Error starting THcareChaincode: %s", err)
	}
}
