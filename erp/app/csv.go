package app

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Define the structure for safe, concurrent CSV writing.
type SyncCSVWriter struct {
	mu       sync.Mutex  // Mutex to synchronize access to the file/writer
	file     *os.File    // Pointer to the open file
	writer   *csv.Writer // CSV writer instance
	header   []string    // The header row to write if the file is new
	filePath string      // Path to the file
}

func NewSyncCSVWriter(filePath string, header []string) (*SyncCSVWriter, error) {
	// Use os.OpenFile with flags:
	// os.O_CREATE: Create the file if it doesn't exist.
	// os.O_APPEND: Append data to the end of the file.
	// os.O_WRONLY: Open the file for writing only.
	// We use the full path to avoid issues with working directories.
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		file, err = os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s: %w", filePath, err)
		}
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
		}
	}

	csvWriter := &SyncCSVWriter{
		file:     file,
		writer:   csv.NewWriter(file),
		header:   header,
		filePath: filePath,
	}

	// Check if the file is empty and write the header if necessary
	if err := csvWriter.writeHeaderIfEmpty(); err != nil {
		err := file.Close()
		if err != nil {
			return nil, err
		} // Close file on error
		return nil, fmt.Errorf("failed to check/write header: %w", err)
	}

	return csvWriter, nil
}

// writeHeaderIfEmpty checks the file size and writes the header if the file is empty.
func (w *SyncCSVWriter) writeHeaderIfEmpty() error {
	w.mu.Lock()
	defer w.mu.Unlock() // Always unlock the mutex when done

	// Get file information
	stat, err := w.file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// If the file size is 0 bytes, it's a new file, so write the header
	if stat.Size() == 0 {
		if err := w.writer.Write(w.header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
		// Flush immediately to ensure header is written before subsequent writes
		w.writer.Flush()
		if err := w.writer.Error(); err != nil {
			return fmt.Errorf("flush error after header write: %w", err)
		}
		fmt.Printf("Header written to %s.\n", w.filePath)
	} else {
		fmt.Printf("File %s already exists and contains data, skipping header write.\n", w.filePath)
	}

	return nil
}

// WriteRow is the synchronized method to write a single row to the CSV file.
func (w *SyncCSVWriter) WriteRow(row []string) error {
	w.mu.Lock()
	defer w.mu.Unlock() // Ensure mutex is released after writing

	if err := w.writer.Write(row); err != nil {
		return fmt.Errorf("failed to write row: %w", err)
	}

	// Flush periodically or after a batch to minimize system calls while ensuring data is written.
	// For this simple example, we'll flush on every write for strong consistency.
	w.writer.Flush()
	if err := w.writer.Error(); err != nil {
		return fmt.Errorf("flush error after row write: %w", err)
	}

	return nil
}

func ReadCSVAndMap() ([]CsvAttendanceLog, error) {
	// 1. Mở file CSV
	f, err := os.Open(CsvPath)
	if err != nil {
		return nil, fmt.Errorf("không thể mở file CSV: %w", err)
	}
	defer f.Close() // Đảm bảo file được đóng khi hàm kết thúc

	// Tạo CSV Reader
	r := csv.NewReader(f)

	// Đọc dòng tiêu đề (header) để bỏ qua
	_, err = r.Read()
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("không thể đọc dòng tiêu đề: %w", err)
	}

	var logs []CsvAttendanceLog // Slice để chứa các object CsvAttendanceLog

	// 2. Lặp qua từng bản ghi (dòng) trong file
	for {
		record, err := r.Read()
		if err == io.EOF {
			break // Đã đọc hết file
		}
		if err != nil {
			return nil, fmt.Errorf("lỗi khi đọc bản ghi: %w", err)
		}

		// Kiểm tra xem dòng có đủ 5 cột không
		if len(record) < 5 {
			fmt.Printf("Bỏ qua dòng không đủ cột: %v\n", record)
			continue
		}

		// 3. Parse chuỗi thời gian
		actionTime, timeErr := time.Parse(TimeLayout, record[2])
		fmt.Println(record[2])
		if timeErr != nil {
			// Xử lý lỗi nếu không parse được thời gian
			fmt.Printf("Lỗi parse thời gian cho dòng %v: %v. Dùng time.Time zero value.\n", record, timeErr)
			actionTime = time.Time{} // Sử dụng giá trị zero nếu có lỗi
		}

		// 4. Map dữ liệu vào struct
		log := CsvAttendanceLog{
			Username:    record[0],
			Action:      record[1],
			ActionTime:  actionTime,
			ErrorDetail: record[3],
			Status:      record[4],
		}

		logs = append(logs, log)
	}

	return logs, nil
}
