package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/nagisa599/LangChain-go-study/constants"
	"github.com/nagisa599/LangChain-go-study/utils"
	"github.com/sashabaranov/go-openai"
)

// チャンクに分割する関数
type ChunkSim struct {
	Index      int
	Similarity float64
}
func main() {
	client := openai.NewClient(os.Getenv("OPENAIAPIKEY"))
	question := "現在のB4のメンバー教えて？"
	fmt.Print("質問: ", question, "\n")
	chunkSize := 400
	overlap := 50
	consolidatedText, err := utils.FetchAndProcessMultipleURLs(constants.Urls)
	if err != nil {
		log.Fatal(err)
	}

	// テキストをチャンクに分割
	chunks := utils.ChunkText(consolidatedText, chunkSize, overlap)
	
	chunksVector ,err := utils.GetEmbedding(client, chunks)
	if err != nil {
		fmt.Println("error")
	}
	questionVector, err := utils.GetEmbedding(client, []string{question})
	if err != nil {
		fmt.Println("error")
	}
	if len(questionVector) == 0 || len(chunksVector) == 0 {
		fmt.Println("Error: chunks vector or question vector is empty.")
		return
	}
	var similarities []ChunkSim 
	for i, vec := range chunksVector {
		similarity, err := utils.CosSimilarity(vec, questionVector[0])
		if err != nil {
			fmt.Printf("Error calculating similarity for chunk %d: %v\n", i, err)
			continue
		}
		similarities = append(similarities, ChunkSim{i, similarity})
	}

	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Similarity > similarities[j].Similarity
	})
	prompt := fmt.Sprintf(`以下の質問に以下の情報をベースにして回答してください。
	[ユーザの情報]
	%s

	[情報]
	%s
	%s
	`, question, chunks[similarities[0].Index], chunks[similarities[1].Index])


	gptChatResponse, err := client.CreateCompletion(context.Background(), openai.CompletionRequest{
		Model:     "gpt-3.5-turbo-instruct", // GPT-3.5-turbo-instructモデルを指定
		Prompt:    prompt,
		MaxTokens: 300, // 応答の最大トークン数
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// レスポンスのテキストを取得し、改行をスペースに置換して余分な空白を削除
	responseText := gptChatResponse.Choices[0].Text
	responseText = strings.ReplaceAll(responseText, "\n", " ") // 改行をスペースに置き換え
	responseText = strings.Join(strings.Fields(responseText), " ") // 余分なスペースを削除

	fmt.Println("GPTの回答:", responseText)

}