package gslog

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"
)

type Formatter interface {
	Format(EntryDTO) string
}

var bf = &basicFormatter{}

type basicFormatter struct{}

func (bf *basicFormatter) Format(e EntryDTO) string {
	if e.Message == "" && len(e.Context) == 0 {
		return ""
	}
	s := e.Time.Format(time.DateTime)
	s += fmt.Sprintf(" [%v]", e.Level)
	s += fmt.Sprintf(" %v", e.Message)
	keys := []string{}
	hasCtx := len(e.Context) > 0
	for k := range e.Context {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	if hasCtx {
		s += ": "
	}
	for _, key := range keys {
		s += fmt.Sprintf(`%v="%v"; `, key, e.Context[key])
	}
	s = strings.TrimSuffix(s, "; ")
	if hasCtx {
		s += ""
	}
	return s
}

func ReadStream(r io.Reader, f Formatter, out chan<- string) {
	defer close(out)
	dec := json.NewDecoder(r)

	for {
		var d EntryDTO
		if err := dec.Decode(&d); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "failed to decode: %v\n", err)
			break
		}

		out <- f.Format(d)
	}
}

func ProcessFile(path string, f Formatter) ([]string, error) {
	if f == nil {
		f = bf
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	dec := json.NewDecoder(reader)

	results := make([]string, 0, 1024)

	for {
		var e EntryDTO

		err := dec.Decode(&e)
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
			// return results, fmt.Errorf("error at offset %d: %w", dec.InputOffset(), err)
		}

		results = append(results, f.Format(e))
	}

	return results, nil
}

func ExtractStructures(r io.Reader, f Formatter) []string {
	if f == nil {
		f = bf
	}
	var results []string

	// Используем Scanner с кастомной логикой поиска блоков { }
	scanner := bufio.NewScanner(r)
	scanner.Split(splitJsonObjects)

	for scanner.Scan() {
		// Получили потенциальный JSON-объект
		data := scanner.Bytes()

		var obj EntryDTO
		// Пробуем десериализовать
		if err := json.Unmarshal(data, &obj); err == nil {
			// Проверка на "пустышку": если в Data обязательные поля,
			// лучше проверить, что они заполнились (например, ID != 0)
			if obj.Message != "" || obj.Level != "" {
				results = append(results, f.Format(obj))
			}
		}
	}
	return results
}

// splitJsonObjects ищет куски текста, начинающиеся на { и заканчивающиеся на }
func splitJsonObjects(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Ищем начало объекта
	start := bytes.IndexByte(data, '{')
	if start == -1 {
		// Если не нашли {, пропускаем всё проверенное, кроме последнего байта (вдруг { будет следующим)
		return len(data), nil, nil
	}

	// Ищем конец объекта (учитываем вложенность)
	depth := 0
	for i := start; i < len(data); i++ {
		if data[i] == '{' {
			depth++
		} else if data[i] == '}' {
			depth--
			if depth == 0 {
				// Мы нашли закрывающую скобку для нашей первой открывающей
				return i + 1, data[start : i+1], nil
			}
		}
	}

	// Если мы здесь, значит нашли { но не нашли }. Ждем больше данных.
	if atEOF {
		return len(data), nil, nil
	}
	return start, nil, nil
}
