package main

import (
	"fmt"
	"io"
	"os"
)

const (
	newDirName = "newFiles"
	oldDirName = "oldFiles"
)

func main() {
	err := CopyFiles(oldDirName, newDirName)
	if err != nil {
		fmt.Println(err)
	}
}

func CopyFiles(oldDir, newDir string) error {
	// получаем список всего содержимого директория
	files, err := os.ReadDir(oldDir)
	if err != nil {
		return err
	}

	// проходимся по всем файлам
	for _, file := range files {
		// проверяем файл на папку
		if file.IsDir() {
			// если да, копируем директорию
			err := CopyDir(oldDir, newDir, file)
			if err != nil {
				return err
			}
		} else {
			// если нет, то копируем файл
			err := CopyFile(oldDir, newDir, file)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CopyFile(oldDir, newDir string, file os.DirEntry) error {
	// если нет, то открываем файл
	content, err := os.Open(oldDir + string(os.PathSeparator) + file.Name())
	if err != nil {
		return err
	}
	defer content.Close()

	// создаем новый файл
	newFile, err := os.Create(newDir + string(os.PathSeparator) + file.Name())
	if err != nil {
		return err
	}
	defer newFile.Close()

	// копируем содержимое
	_, err = io.Copy(newFile, content)
	if err != nil {
		return err
	}
	return nil
}

func CopyDir(oldDir, newDir string, file os.DirEntry) error {
	err := os.Mkdir(newDir+string(os.PathSeparator)+file.Name(), os.ModePerm)
	if err != nil {
		return err
	}
	// если да, то рекурсивно вызываем функцию
	err = CopyFiles(oldDir+string(os.PathSeparator)+file.Name(), newDir+string(os.PathSeparator)+file.Name())
	if err != nil {
		return err
	}
	return nil
}
