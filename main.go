// 2022-09-11
// https://github.com/DistributedMetaverse/distributed-metaverse-offchain
// 공개SW대회용 오프체인(Offchain) 구현'
// Go Namhyeon <gnh1201@gmail.com>

package main

import (
    "os"
    "io"
    "os/exec"
    "context"
    "fmt"
    "time"
    "bytes"
    "strings"
    "strconv"
    "net/http"
    "encoding/json"
    "math/rand"
    "crypto/sha256"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var rdb *redis.Client
var storedTransactions []Transaction

type Block struct {
    PreviousHash string `json:"previousHash"`
    Transactions []Transaction `json:"transactions"`
    Hash string `json:"hash"`
    Datetime string `json:"datetime"`
    Proof int `json:"proof"`
    LastTransactionId int `json:"lastTransactionId"`
}

type Transaction struct {
    Data string `json:"data"`
    Id int `json:"id"`
    Datetime string `json:"datetime"`
}

type IPFSTransactionData struct {
    QmHash string `json:"qmhash"`
    MIMEType string `json:"mimetype"`
    Filename string `json:"filename"`
}

// 해시 계산
func (b *Block) CalculateHash() {
    jsonBytes, _ := json.Marshal(b.Transactions)
    blockData := b.PreviousHash + string(jsonBytes) + b.Datetime + strconv.Itoa(b.Proof)
    blockHash := sha256.Sum256([]byte(blockData))
    b.Hash = fmt.Sprintf("%x", blockHash)
}

// 작업 증명 (PoW)
func (b *Block) pow(difficulty int) {
    for !strings.HasPrefix(b.Hash, strings.Repeat("0", difficulty)) {
        b.Proof++
        b.CalculateHash()
    }
}

// 외부 API
func serve() {
    // Echo instance
    e := echo.New()

    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    // Routes
    e.POST("/transaction/publish", publishTransaction)
    e.GET("/transaction/:id", queryTransaction)
    e.GET("/block/:hash", getBlockInfo)
    e.GET("/chain/:depth", getLastBlocks)
    e.GET("/stat", getStat)
    e.POST("/upload", uploadFile)
    e.Static("/downloads", "downloads")

    // Start server
    e.Logger.Fatal(e.Start(":1323"))
}

// 트랜젝션 수신 (고루틴)
func runReceiveTransactions() {
    pubsub := rdb.Subscribe(ctx, "transactions_live")
    defer pubsub.Close()

    // 메시지 계속 수신
    for {
        // 최근 메시지 수신
        msg, err := pubsub.ReceiveMessage(ctx)
        if err != nil {
            continue
        }

        transaction := Transaction{}   // 트랜젝션 객체 생성
        json.Unmarshal([]byte(msg.Payload), &transaction)   // JSON을 객체로 변환
        storedTransactions = append(storedTransactions, transaction)    // 트랜젝션 추가

        // 콘솔 메시지 출력
        fmt.Println(msg.Channel, msg.Payload)
    }
}

// 블록 수신 (고루틴)
func runReceiveBlocks() {
    pubsub := rdb.Subscribe(ctx, "blocks_live")
    defer pubsub.Close()

    // 메시지 계속 수신
    for {
        // 최근 메시지 수신
        msg, err := pubsub.ReceiveMessage(ctx)
        if err != nil {
            continue
        }
        
        block := Block{}   // 블록 객체 생성
        json.Unmarshal([]byte(msg.Payload), &block)   // JSON을 객체로 변환
        
        // 발행할 블록 본문 만들기
        jsonBytes, err := json.Marshal(block)
        if err != nil {
            continue
        }

        // 블록을 로컬에 파일로 저장
        saveBlock(block.Hash, jsonBytes)
        
        // 콘솔 메시지 출력
        fmt.Println(msg.Channel, msg.Payload)
    }
}

// 작업 증명(PoW) 실행 (고루틴)
func runProof() {
    for {
        // 블록 생성
        block, err := createBlock()
        if err != nil {
            continue
        }
        fmt.Printf("New block! %s\n", block.Hash)
    }
}

// 블록 빌드
func buildBlock() (Block, error) {
    var transactions []Transaction   // 트랜젝션 변수 정의

    // 마지막 트랜젝션 ID 구하기
    lastTransactionId, _ := getLastTransactionId()
    
    // 마지막 블록 Hash 구하기
    lastBlockHash, _ := getLastBlockHash()

    // 현재 트랜젝션 복사
    //_storedTransactions := make([]Transaction, len(storedTransactions))
    //copy(storedTransactions, _storedTransactions)
    //storedTransactions = make([]Transaction, 0)

    // 트랜젝션 확인
    for _, transaction := range storedTransactions {
        if transaction.Id <= lastTransactionId {
            transactions = append(transactions, transaction)
        }
    }
    storedTransactions = nil   // 메모리에 남은 모든 트랜젝션 삭제

    // 블록 생성
    block := Block{
        PreviousHash: lastBlockHash,
        Transactions: transactions,
        Datetime: time.Now().String(),
        Proof: 0,
        LastTransactionId: lastTransactionId,
    }

    // 블록 해시 계산
    block.CalculateHash()

    // 블록 반환
    return block, nil
}

// 트랜젝션 빌드
func buildTransaction(data string) (Transaction, error) {
    lastTransactionId, err := getLastTransactionId()

    if err != nil {
        return Transaction{}, err
    }

    transaction := Transaction{
        Data: data,
        Id: lastTransactionId + 1,
        Datetime: time.Now().String(),
    }

    return transaction, nil
}

// 최근 트랜젝션 ID 구하기
func getLastTransactionId() (int, error) {
    lastTransactionId := -1

    result, _ := rdb.Get(ctx, "lastTransactionId").Result()
    transactionId, err := strconv.Atoi(result)

    if err != nil {
        err := rdb.Set(ctx, "lastTransactionId", "0", 0).Err()
        if err != nil {
            lastTransactionId = 0
        }
    } else {
        lastTransactionId = transactionId
    }

    return lastTransactionId, err
}

// 최근 블록 수 구하기
func getLastBlocksCount() (int, error) {
	lastBlocksCount := -1

    result, _ := rdb.Get(ctx, "lastBlocksCount").Result()
    blocksCount, err := strconv.Atoi(result)

    if err != nil {
        err := rdb.Set(ctx, "lastBlocksCount", "0", 0).Err()
        if err != nil {
            lastBlocksCount = 0
        }
    } else {
        lastBlocksCount = blocksCount
    }

    return lastBlocksCount, err
}

// 최근 블록 해시 구하기
func getLastBlockHash() (string, error) {
    lastBlockHash := ""

    result, err := rdb.Get(ctx, "lastBlockHash").Result()
    if err != nil {
        //block, err := createBlock()
        //return block.hash, err
        return lastBlockHash, err
    } else {
        lastBlockHash = result
    }

    return lastBlockHash, nil
}

// 시작점(Entrypoint)
func main() {
    // Redis 서버에 연결
    rdb = redis.NewClient(&redis.Options{
        Addr:     "154.12.242.48:60713",
        //Password: "", // no password set
        //DB:       0,  // use default DB
    })
    
    // 최근 트랜젝션 ID 확인
    lastTransactionId, _ := getLastTransactionId()
    fmt.Printf("lastTransactionId: %d\n", lastTransactionId)
    
    // 최근 블록 해시 확인
    lastBlockHash, _ := getLastBlockHash()
    fmt.Printf("lastBlockHash: %s\n", lastBlockHash)

    // 트랜젝션 수신 루틴 실행
    go runReceiveTransactions()

    // 블록 수신 루틴 실행
    go runReceiveBlocks()

    // 작업 증명 루틴 실행
    go runProof()

    // 웹서버 실행
    serve()
}

// 블록 생성
func createBlock() (Block, error) {
    block, err := buildBlock()

    // 마지막 블록 Hash 구하기
    lastBlockHash, _ := getLastBlockHash()
    
    // 작업 증명(PoW) 실행
    block.pow(6)    // PoW 문제를 풀면 블록 생성 (채굴과 동일함, CPU 사용률 높음)
    //time.Sleep(5 * time.Minute)     // 5분마다 1개씩 블록 생성 (CPU 사용량 낮음)

    // 마지막 블록 Hash 다시 구하기
    _lastBlockHash, _ := getLastBlockHash()

    // 작업 증명이 이루어지는 동안 다른 블록이 생성된 경우 (먼저 블록을 찾은 노드가 승리)
    if lastBlockHash != _lastBlockHash {
        err = fmt.Errorf("Lost Block! %s", block.Hash)
        return block, err
    }

    // 발행할 블록 본문 만들기
    jsonBytes, err := json.Marshal(block)
    if err != nil {
        return block, err
    }

    // 블록을 Redis에 저장
    err2 := rdb.Set(ctx, block.Hash, string(jsonBytes), 0).Err()
    if err2 != nil {
        return block, err2
    }

    // 최근 해시 갱신
    err3 := rdb.Set(ctx, "lastBlockHash", block.Hash, 0).Err()
    if err3 != nil {
        return block, err3
    }
    
    // 블록을 Redis에 발행
    err4 := rdb.Publish(ctx, "blocks_live", string(jsonBytes)).Err()
    if err4 != nil {
        return block, err4
    }

    // 발생된 블록을 로컬에 파일로 저장
    saveBlock(block.Hash, jsonBytes)

    return block, err
}

// 디렉토리 존재 여부
func isExists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return false, err
}

// 블록을 로컬에 저장
func saveBlock(hash string, jsonBytes []byte) error {
    path := "data"   // 파일 디렉토리 경로 지정

    // 폴더 확인 및 생성
    result, err := isExists(path)
    if result == false || err != nil {
        err := os.Mkdir(path, os.ModePerm)
        if err != nil {
            return err
        }
    }

    // 파일이 존재하면 넘어감
    path2 := "data/" + hash + ".json"
    result2, err := isExists(path2)
    if !(result2 == false || err != nil) {
        return nil
    }

    // 블록을 저장할 파일 생성
    f, err := os.Create(path2)
    defer f.Close()    // 작업을 완료하면 파일을 닫음
    if err != nil {
        return err
    }
    
    // 파일에 내용 작성
    n, err := f.Write(jsonBytes)
    if err != nil {
        return err
    }
    fmt.Printf("wrote %d bytes\n", n)

    return nil
}

// 트랜젝션 발행
func publishTransaction(c echo.Context) error {
    // 받은 요청 해석
    jsonMap := make(map[string]interface{})
    err := json.NewDecoder(c.Request().Body).Decode(&jsonMap)
    if err != nil {
        return err
    }

    // 변수로 재정렬
    transactionData := IPFSTransactionData{
        QmHash: jsonMap["qmhash"].(string),
        MIMEType: jsonMap["mimetype"].(string),
        Filename: jsonMap["filename"].(string),
    }
    jsonBytes_transactionData, err := json.Marshal(transactionData)

    // 신규 트랜젝션 생성
    transaction, err := buildTransaction(string(jsonBytes_transactionData))
    if err != nil {
        return err
    }

    // 발행할 트랜젝션 본문 만들기
    jsonBytes_transaction, err := json.Marshal(transaction)
    if err != nil {
        return err
    }

    // 마지막 트랜젝션 ID 올리기
    err2 := rdb.Set(ctx, "lastTransactionId", strconv.Itoa(transaction.Id), 0).Err()
    if err2 != nil {
        return err2
    }

    // 트랜젝션 발행
    err3 := rdb.Publish(ctx, "transactions_live", string(jsonBytes_transaction)).Err()
    if err3 != nil {
        return err3
    }

    // 모든 작업이 완료되었으면 오류 없음으로 반환
    response := map[string]interface{}{
        "success": true,
        "id": transaction.Id,
    }
    return c.JSON(http.StatusOK, response)
}

// 블록 정보 불러오기
func loadBlock(hash string) (Block, error) {
    // 블록 객체 생성
    block := Block{}   

    // 로컬에서 블록 불러오기
    b, err := os.ReadFile("data/" + hash + ".json")
    if err != nil {
        // 블록이 로컬에 없는 경우 Redis에서 조회
        result, err := rdb.Get(ctx, hash).Result()
        if err != nil {
            return block, err
        } else {
            b = []byte(result)
        }
    }

    // JSON을 객체로 변환
    json.Unmarshal(b, &block)  

    // 블록 반환
    return block, err
}

// 블록 정보조회
func getBlockInfo(c echo.Context) error {
    hash := c.Param("hash")   // Hash 값 수신
    
    // 블록 정보 불러오기
    block, err := loadBlock(hash)
    if err != nil {
        return err
    }

    // 모든 작업이 완료되었으면 오류 없음으로 반환
    response := map[string]interface{}{
        "success": true,
        "data": block,
    }
    return c.JSON(http.StatusOK, response)
}

// 블록 기록 조회 (최근 블록 기준)
func getLastBlocks(c echo.Context) error {
    var blocks []Block   // 블록 정의

    depth := c.Param("depth")   // 탐색 깊이
    n, err := strconv.Atoi(depth)   // 숫자로 변환

    // 마지막 블록 Hash 구하기
    lastBlockHash, err := getLastBlockHash()
    if err != nil {
        return err
    }

    // 블록 탐색
    blockHash := lastBlockHash
    for n > 0 {
        // 블록을 찾고 추가
        block, err := loadBlock(blockHash)
        if err != nil {
            break
        }
        blocks = append(blocks, block)
        
        // 이전 블록 탐색
        if block.PreviousHash != "" {
            blockHash = block.PreviousHash
        }
        
        // 탐색 횟수 업데이트
        n--
    }
    
    // 모든 작업이 완료되었으면 오류 없음으로 반환
    response := map[string]interface{}{
        "success": true,
        "data": blocks,
    }
    return c.JSON(http.StatusOK, response)
}

// 트랜젝션 ID로 트랜젝션 찾기
func searchByTransactionId(transactionId int) (Transaction, error) {
    var transaction Transaction   // 트랜젝션

    // 마지막 블록 Hash 구하기
    lastBlockHash, err := getLastBlockHash()
    if err != nil {
        return transaction, err
    }

    // 블록 탐색
    blockHash := lastBlockHash
    for {
        // 블록 조회
        block, err := loadBlock(blockHash)
        if err != nil {
            break
        }
        
        // 트랜젝션 찾기
        for _, _transaction := range block.Transactions {
            if _transaction.Id == transactionId {
                transaction = _transaction
                break
            }
        }

        // 찾으려는 트랜젝션 ID가 블록의 마지막 트랜젝션 ID보다 번호가 낮은 경우 (찾을 수 없는 경우)
        if block.LastTransactionId < transactionId {
            err = fmt.Errorf("%s", "Transaction not found")
            break
        }

        // 이전 블록 탐색
        if block.PreviousHash != "" {
            blockHash = block.PreviousHash
        } else {
            err = fmt.Errorf("%s", "Transaction not found")
            break
        }
    }

    return transaction, err
}

// 트랜젝션 퀴리
func queryTransaction(c echo.Context) error {
    transactionId := c.Param("id")   // 트랜젝션 ID 받기

    // 트랜젝션 확인
    n, _ := strconv.Atoi(transactionId)
    transaction, err := searchByTransactionId(n)
    if err != nil {
        return err
    }

    // 트랜젝션 해석
    transactionData := IPFSTransactionData{}
    json.Unmarshal([]byte(transaction.Data), &transactionData)

    // 다운로드 폴더 확인 및 생성
    downloadPath := "downloads"
    result, err := isExists(downloadPath)
    if result == false || err != nil {
        err := os.Mkdir(downloadPath, os.ModePerm)
        if err != nil {
            return err
        }
    }

    // 파일 존재 확인
    filePath :=  downloadPath + "/" + transactionData.QmHash
    result2, err := isExists(filePath)
    
    // 없으면 다운로드 시작
    if result2 == false || err != nil {
        cmd := exec.Command("../kubo/ipfs", "get", transactionData.QmHash)
        cmd.Dir = downloadPath
        cmd.Start()   // 완료를 기다리지 않음

        response := map[string]interface{}{
            "success": true,
            "data": transaction,
            "status": "downloading",
        }
        return c.JSON(http.StatusOK, response)
    }

    // 있으면 파일 주소로 응답
    response := map[string]interface{}{
        "success": true,
        "data": transaction,
        "status": "ok",
        "url": "http://127.0.0.1:1323/downloads/" + transactionData.QmHash,
    }
    return c.JSON(http.StatusOK, response)
}

// 파일 업로드
func uploadFile(c echo.Context) error {
    // 다운로드 폴더 확인 및 생성
    uploadPath := "uploads"
    result, err := isExists(uploadPath)
    if result == false || err != nil {
        err := os.Mkdir(uploadPath, os.ModePerm)
        if err != nil {
            return err
        }
    }

    // Source
    file, err := c.FormFile("file")
    if err != nil {
        return err
    }
    src, err := file.Open()
    if err != nil {
        return err
    }
    defer src.Close()
    
    // Destination
    filename := strconv.Itoa(rand.Intn(1000000))
    filepath := uploadPath + "/" + filename
    dst, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer dst.Close()

    // Copy
    if _, err = io.Copy(dst, src); err != nil {
        return err
    }

    // IPFS에 업로드
    var bOut, bErr bytes.Buffer
    cmd := exec.Command("../kubo/ipfs", "add", filename)
    cmd.Dir = uploadPath
    cmd.Stdout = &bOut
    cmd.Stderr = &bErr
    cmd.Run()

    // 업로드 완료 체크
    s := bOut.String()
    qmHash := ""
    pos := 0
    for pos > -1 {
        pos = strings.Index(s, " ")

        if strings.HasPrefix(s, "Qm") {
            qmHash = s[0:pos]
            break
        }

        s = s[pos+1:]
    }

    // QmHash 파싱에 실패한 경우
    if qmHash == "" {
        return fmt.Errorf("%s", "QmHash not found")
    }

    // 모든 작업이 완료되었으면
    response := map[string]interface{}{
        "success": true,
        "qmhash": qmHash,
    }
    return c.JSON(http.StatusOK, response)
}

func getStat(c echo.Context) error {
	lastTransactionId, _ := getLastTransactionId()
	lastBlockHash, _ := getLastBlockHash()
	lastBlocksCount, _ := getLastBlocksCount()

    response := map[string]interface{}{
        "lastTransactionId": lastTransactionId,
        "lastBlockHash": lastBlockHash,
		"lastBlocksCount": lastBlocksCount,
    }
    return c.JSON(http.StatusOK, response)
}

// References:
//     https://gist.github.com/LordGhostX/bb92b907731ee8ebe465a28c5c431cb4
//     https://redis.uptrace.dev/guide/go-redis-pubsub.html
//     https://stackoverflow.com/questions/41410655/extract-json-from-golangs-echo-request
//     https://stackoverflow.com/questions/27137521/how-to-convert-interface-to-string
//     https://stackoverflow.com/questions/26327391/json-marshalstruct-returns
//     https://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-exists
//     https://echo.labstack.com/guide/request/
//     https://echo.labstack.com/cookbook/file-upload/
//     https://stackoverflow.com/questions/1877045/how-do-you-get-the-output-of-a-system-command-in-go
//     https://pkg.go.dev/time
