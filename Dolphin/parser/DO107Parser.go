/*
	This is a sample code for decoding bytes of DO107 data packet, coded in the Go language. To use these codes, you must put the bytes received in the receivedBytes variable  and leave the hexString variable blank  or put the received hex string in the hexString variable.

*/
package parser

import (
	"bufio"
	"strings"

	//"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"strconv"
	//"strings"

	proto "github.com/golang/protobuf/proto"
	"github.com/pierrec/lz4"
	"io/ioutil"
	"net/http"
)


const (
	GZIP_FLAG = 0x00000001
)

func Parse() {
	receivedBytes := []byte{}
	hexString := "ababe80300000001000001003a0d13d44f1603003f0f0000b4099da60a410d9bb80242159f9c4e421df3faca44200e2d59c0043e35d0401543380745d74cbf6048015a0a0a0b1e32333435363764620f000014d4f03615b39306000000da010a410d9bb8024215aa9c4e421d347aca44200c2db532a13e353000f5423808452f4fbf6048015a0a0a0b1e32333435363764620f000000d8f03604b39306000000c4010a410da5b8024215c99c4e421d4c62ce44200d2dc989f63e35983f42433807458851bf6048015a0a0a0b1e32333435363764620f000015e5f0360db3930600000086020a410d9ab8024215759c4e421d8171cc4420102d2aaeaa3e354d032743380645e153bf6048015a0a0a0b1e32333435363764620f00001583f1361eb3930600000086020a410d25b9024215319d4e421d9aa9d44420062d2aaeaa3f355238af423812452054bf6048015a0a0a0b1e32333435363764620f00001283f13600b3930600000086020a410d17b9024215199d4e421d9105d34420112dfa9b503e355238af423807456e56bf6048015a0a0a0b1e32333435363764620f0000158ef1360bb3930600000091020a410d17b9024215199d4e421d9105d34420122d59c0043e355238af42380745c758bf6048015a0a0a0b1e32333435363764620f0000148ef13600b393060000009c020a410d17b9024215199d4e421d9105d34420112d59c0043e355238af42380645205bbf6048015a0a0a0b1e32333435363764620f0000148ef13600b3930600000091020a410d17b9024215199d4e421d9105d34420112d8520c73e355238af42380745785dbf6048015a0a0a0b1e32333435363764620f0000148ef13600b393060000009c020a410dc3b80242157e9c4e421d9ac9ce4420092d3546bb40359ab9b243380e45bc5dbf6048015a0a0a0b1e32333435363764620f0000138ef13600b39306000000cf010a410dc5b8024215859c4e421d3060ce4420122dfa9bd03f35ba938243380745d35ebf60480a5a0a0a0b1e32333435363764620f01011491f13603b493060100009c020a410dc5b8024215859c4e421d3060ce4420122d8520c73e35ba938243380745d55ebf60480a5a0a0a0b1e32333435363764620f00001491f13600b5930601000086020a410dc5b8024215859c4e421d3060ce4420122d7efb724035ba938243380745d65ebf60480a5a0a0a0b1e32333435363764620f01011491f13600b6930601000091020a410d7fb8024215889c4e421d66c6c84420122d8221f74035cd2cb243380745d75ebf6048015a0a0a0b1e32333435363764620f010114aff1361eb7930601000086020a420d4db802421506a14e421d3323c64420112d3cc2e04135d723af42380745135fbf6048015a0a0a0b1e323334353637646210010108d2f436"
	var err error
	// Conver Hex String to byte array
	if hexString != "" {
		receivedBytes, _ = hex.DecodeString(hexString)
	}
	buffer := bytes.NewBuffer(receivedBytes)
	buffer.Write(receivedBytes[:len(receivedBytes)])

	expectedSize := int32(0)
	flag := uint16(0)
	header := []byte{}
	var crc uint32
	var messageType int32
	for {

		if expectedSize == 0 && buffer.Len() >= 28 {
			//skip header data= ABAB
			header = buffer.Bytes()[:24]
			buffer.Next(2)

			fmt.Println("Tid", binary.LittleEndian.Uint32(buffer.Next(4)))

			fmt.Println("API Version", binary.BigEndian.Uint16(buffer.Next(2)))

			flag = binary.BigEndian.Uint16(buffer.Next(2))
			fmt.Println("Flag", flag)

			messageType = int32(binary.LittleEndian.Uint16(buffer.Next(2)))
			fmt.Println("Message Type", MessageType_name[messageType])

			fmt.Println("IMEI", binary.LittleEndian.Uint64(buffer.Next(8)))

			expectedSize = int32(binary.LittleEndian.Uint32(buffer.Next(4)))

			fmt.Println("Expected size", expectedSize)
			crc = binary.LittleEndian.Uint32(buffer.Next(4))
			fmt.Println("CRC", crc)

		}

		if buffer.Len() >= int(expectedSize) && buffer.Len() != 0 {

			fmt.Println("Remining buffer length", buffer.Len())
			data := buffer.Next(int(expectedSize))

			if crc != 0 {
				packetCRC := crc32.Update(0, crc32.MakeTable(crc32.Castagnoli), header)
				if expectedSize > 0 {
					packetCRC = crc32.Update(packetCRC, crc32.MakeTable(crc32.Castagnoli), data)
				}

				if packetCRC != crc {
					panic("CRC Error,Generated CRC:" + strconv.FormatUint(uint64(packetCRC), 10))
				}

			}

			fmt.Println("Bytes of data ", hex.EncodeToString(data))
			fmt.Println("New buffer length", hex.EncodeToString(buffer.Bytes()))
			// Decompress data

			if (flag&GZIP_FLAG) > 0 && len(data) > 0 {
				out := make([]byte, 4096)

				size, err := lz4.UncompressBlock(data, out)
				if err != nil {
					panic(err.Error())

					return
				}

				data = out[:size]
				fmt.Println("Uncompress block of data with lz4 ", hex.EncodeToString(out))
			}
			// fmt.Println("Uncompress block of data with lz4 ", hex.EncodeToString(data))
			expectedSize = 0

			// Parse data based on message type
			switch MessageType(messageType) {
			case MessageType_DataPack_Request:
				request := &DataPackRequest{}
				err = proto.Unmarshal(data, request)
				if err == nil {
					fmt.Println("Parsed Message", proto.MarshalTextString(request))
				}
				// Call TB API
				httpposturl := "http://192.168.50.26:8080/api/v1/gAy3RYdXEQ43ypkOKAqi/telemetry"
				fmt.Println("HTTP JSON POST URL:", httpposturl)

				var jsonData = []byte(`{}`)

				var m map[string]interface{}
				err := json.Unmarshal(jsonData, &m)
				fmt.Println(err)
				arr := [27][2]string{}
				if err == nil {
					count := 0
					parsed :=  proto.MarshalTextString(request)
					scanner := bufio.NewScanner(strings.NewReader(parsed))
					for scanner.Scan() {
						tr := strings.Replace(scanner.Text()," ", "", -1)
						//temp := scanner.Text()
						if count > 0 && count < 10 {
							key := strings.Split(tr, ":")[0]
							value := strings.Split(tr, ":")[1]
							m[key] = value
						}

						if strings.Split(tr, ":")[0] == "IOList_ID"{
							arr[count-10][0] = strings.Split(tr, ":")[1]
						}
						if strings.Split(tr, ":")[0] == "IOList_Value"{
							arr[count-37][1] = strings.Split(tr, ":")[1]
						}
						count++
					}
				}
				for i := 0; i < 27; i++{
					m[arr[i][0]] = arr[i][1]
				}

				newdata , err := json.Marshal(m)
				fmt.Println("new", newdata)
				req, error := http.NewRequest("POST", httpposturl, bytes.NewBuffer(newdata))
				req.Header.Set("Content-Type", "application/json; charset=UTF-8")

				client := &http.Client{}
				response, error := client.Do(req)
				if error != nil {
					panic(error)
				}
				defer response.Body.Close()

				fmt.Println("response Status:", response.Status)
				fmt.Println("response Headers:", response.Header)
				body, _ := ioutil.ReadAll(response.Body)
				fmt.Println("response Body:", string(body))

			case MessageType_Config_Set_Response:
				request := &ConfigSetResponse{}
				err = proto.Unmarshal(data, request)
				if err == nil {
					fmt.Println("Parsed Message ", proto.MarshalTextString(request))
				}

			case MessageType_Config_Get_Response:
				request := &ConfigGetResponse{}
				err = proto.Unmarshal(data, request)
				if err == nil {
					fmt.Println("Parsed Message ", proto.MarshalTextString(request))
				}

			case MessageType_Firmware_Update_Start_Response, MessageType_Firmware_GetPack_Request, MessageType_Firmware_Update_State:
				request := &FirmwareUpdateResp{}
				err = proto.Unmarshal(data, request)
				if err == nil {
					fmt.Println("Parsed Message ", proto.MarshalTextString(request))
				}

			case MessageType_Log_Pack:
				request := &LogPack{}
				err = proto.Unmarshal(data, request)
				if err == nil {
					fmt.Println("Parsed Message ", proto.MarshalTextString(request))
				}

			case MessageType_Command_Response:
				request := &CommandResponse{}
				err = proto.Unmarshal(data, request)
				if err == nil {
					fmt.Println("Parsed Message ", proto.MarshalTextString(request))
				}

			default:
				fmt.Println("Command not found", messageType)

			}
			if err != nil {
				panic("Decoding field, err: " + err.Error())
			}

		} else {
			fmt.Println("buffer length", buffer.Len())
			return
		}

	}

}
