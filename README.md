# Secretly

## Documentation

TODO

## Usage

## Structure Definition

### Tag Support

```go
type Specification struct {
    // A secret named "TextSecret" that stores text data.
    TextSecret            string `type:"text"`

    // A secret named "TextSecretVersion" that stores text data. Rather than retrieving the latest version, retrieve version 1.
    TextSecretWithVersion1 string `type:"text" version:"1"`

    // A secret named "Split_Text_Secret" that stores text data.
    SplitTextSecret       string `type:"text" split_words:"true"`
    
    // A secret named "Map_Secret" that stores mapped data including a key "Map_Secret_Key".
    MapSecretKey          int `type:"map" secret_name:"Map_Secret" split_words:"true"`

    // A secret named "Map_Secret" that stores mapped data with a key "Map_Secret_Key_2".
    MapSecretExplicitKey  float64 `type:"map" secret_name:"Map_Secret" key_name:"Map_Secret_Key_2"`

    // Ignored.
    IgnoredField          string `ignored:"true"`
}
```

### Supported Field types
