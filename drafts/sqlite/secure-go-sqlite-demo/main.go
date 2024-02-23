
package main

import (
    "database/sql"
    "fmt"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "log"
    "path/filepath"

    _ "github.com/mattn/go-sqlite3" // This import is necessary for SQLite support
)

type Config struct {
    SQLCipherKey struct {
        EncryptionKey string `yaml:"encryption_key"`
    } `yaml:"sqlcipher_key"`
}

// LoadConfig reads the configuration from the given YAML file.
func LoadConfig(path string) (*Config, error) {
    var config Config
    filename, err := filepath.Abs(path)
    if err != nil {
        return nil, err
    }
    yamlFile, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    err = yaml.Unmarshal(yamlFile, &config)
    if err != nil {
        return nil, err
    }
    return &config, nil
}

func main() {
    config, err := LoadConfig("../../ignore/secrets.yaml")
    if err != nil {
        log.Fatalf("Error loading config: %v", err)
    }

    // Convert the base64-encoded key to its raw form if necessary
    encryptionKey := config.SQLCipherKey.EncryptionKey

    // Open the encrypted database using the SQLCipher key
    db, err := sql.Open("sqlite3", fmt.Sprintf("./secure_db.sqlite?_pragma_key=x'%s'&_pragma_cipher_page_size=4096", encryptionKey))
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Perform your database operations here...
    // For example, creating a table, inserting data, querying, etc.
}

