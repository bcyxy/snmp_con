package hdrtree

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
)

//HandlerJSON xxx
type HandlerJSON struct {
	Input  string                     `json:"input"`
	Record map[string]string          `json:"record"`
	Regx   string                     `json:"regx"`
	Next   map[string]json.RawMessage `json:"next"`
}

//HdrNode xxx
type HdrNode struct {
	Input  string
	Record map[string]string
	Regx   *regexp.Regexp
	Next   map[string]HdrNode
}

//LoadFromFile xxx
func (selfPtr *HdrNode) LoadFromFile(fPath string) error {
	jsonText, err := ioutil.ReadFile(fPath)
	if err != nil {
		return err
	}

	var myStruct json.RawMessage
	err = json.Unmarshal(jsonText, &myStruct)
	if err != nil {
		return err
	}
	selfPtr.loadNode(myStruct)

	return nil
}

//loadNode xxx
func (selfPtr *HdrNode) loadNode(myStruct json.RawMessage) error {
	tmpHdJSON := &HandlerJSON{
		Record: make(map[string]string),
		Next:   make(map[string]json.RawMessage),
	}
	err := json.Unmarshal(myStruct, tmpHdJSON)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}
	selfPtr.Input = tmpHdJSON.Input
	selfPtr.Record = tmpHdJSON.Record
	selfPtr.Regx, err = regexp.Compile(tmpHdJSON.Regx)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	for key, hdrStr := range tmpHdJSON.Next {
		subHdr := HdrNode{
			Record: make(map[string]string),
			Next:   make(map[string]HdrNode),
		}
		subHdr.loadNode(hdrStr)
		selfPtr.Next[key] = subHdr
	}
	return nil
}

//GetVals xxx
func (selfPtr *HdrNode) GetVals(devInfo, hdrKey, preInput string, record map[string]string) error {
	//input
	var input string
	var err error
	if strings.HasPrefix(selfPtr.Input, "OID:") {
		oid := selfPtr.Input[4:len(selfPtr.Input)]
		input, err = getOidSingleStr(devInfo, oid)
		if err != nil {
			return err
		}
	} else if selfPtr.Input == "$PRE_INPUT" {
		input = preInput
	}

	//handle
	var reRstStr string
	reRst := selfPtr.Regx.FindAllStringSubmatch(input, 1)
	if len(reRst) > 0 {
		firstRst := reRst[0]
		reRstStr = firstRst[len(firstRst)-1]
	}

	//record
	for rcdKey, rcdType := range selfPtr.Record {
		if !strings.HasPrefix(rcdType, "$") {
			record[rcdKey] = rcdType
			continue
		}

		switch rcdType {
		case "$RE_RST":
			record[rcdKey] = reRstStr
		case "$KEY":
			record[rcdKey] = hdrKey
		}
	}

	//next
	nextNode, ok := selfPtr.Next[reRstStr]
	if !ok {
		nextNode, ok = selfPtr.Next["*"]
		if !ok {
			return nil
		}
	}
	return nextNode.GetVals(devInfo, reRstStr, input, record)
}

func getOidSingleStr(ip, oid string) (string, error) {
	// SNMP参数
	snmpConn := &gosnmp.GoSNMP{
		Target:    ip,
		Port:      161,
		Community: "public", //私有云:netadmin00ip
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(5) * time.Second,
		Retries:   2,
		MaxOids:   gosnmp.MaxOids,
	}

	// 连接
	err := snmpConn.Connect()
	if err != nil {
		return "", err
	}
	defer snmpConn.Conn.Close()

	// snmpwalk
	snmpRes, err := snmpConn.WalkAll(oid)
	if err != nil {
		return "", err
	}
	if len(snmpRes) != 1 {
		return "", fmt.Errorf("oid '%s' return more than one", oid)
	}

	pdu := snmpRes[0]
	var pduValStr string
	switch pdu.Type {
	case gosnmp.OctetString:
		pduValStr = string(pdu.Value.([]byte))
	default:
		return "", fmt.Errorf("pdu type '%v' not support", pdu.Type)
	}

	return pduValStr, nil
}
