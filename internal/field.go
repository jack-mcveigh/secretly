package secretly

import "reflect"

const (
	DefaultType    = "text"
	DefaultVersion = "0"
	TagIgnored     = "ignored"
	TagKeyName     = "key_name"
	TagSecretName  = "secret_name"
	TagSplitWords  = "split_words"
	TagType        = "type"
	TagVersion     = "version"
)

type field struct {
	SecretType    string
	SecretName    string
	SecretVersion string
	MapKeyName    string // NOTE: Only used when secretType is "map"
	Value         reflect.Value
}

func NewField(fValue reflect.Value, fStructField reflect.StructField) (field, error) {
	// reduce pointer to value/struct pointer. Initialize underlying struct if needed
	// TODO: maybe remove struct handling, might not be needed for this implementation
	for fValue.Kind() == reflect.Ptr {
		if fValue.IsNil() {
			if fValue.Type().Elem().Kind() != reflect.Struct {
				// value other than struct
				break
			}
			// value is a struct, initialize it
			fValue.Set(reflect.New(fValue.Type().Elem()))
		}
		fValue = fValue.Elem()
	}

	var newField field

	// Get the split_words value, setting it to false if not explicitly set
	splitWordsEnabled, ok, err := parseOptionalStructTagKey[bool](fStructField, TagSplitWords)
	if err != nil {
		return field{}, StructTagError{
			name: fStructField.Name,
			key:  TagSplitWords,
			err:  err,
		}
	}
	if !ok {
		splitWordsEnabled = false
	}

	// Get the type value, setting it to the default, "text", if not explicitly set.
	// Also perform validation to ensure only valid types are provided
	newField.SecretType, ok, err = parseOptionalStructTagKey[string](fStructField, TagType)
	if err != nil {
		return field{}, StructTagError{
			name: fStructField.Name,
			key:  TagType,
			err:  err,
		}
	}
	if !ok {
		newField.SecretType = DefaultType
	}
	switch newField.SecretType {
	case "text", "map":
	default:
		return field{}, StructTagError{
			name: fStructField.Name,
			key:  TagType,
			err:  ErrInvalidSecretType,
		}
	}

	// Get the secret_name value, setting it to the field's name if not explicitly set.
	// Split the words if the default value was used and split_words was set to true
	newField.SecretName, ok, err = parseOptionalStructTagKey[string](fStructField, TagSecretName)
	if err != nil {
		return field{}, StructTagError{
			name: fStructField.Name,
			key:  TagSecretName,
			err:  err,
		}
	}
	if !ok {
		newField.SecretName = fStructField.Name
		if splitWordsEnabled {
			newField.SecretName = splitWords(newField.SecretName)
		}
	}

	// Get the key_name value, if the type is "map", and setting it to the field's name
	// if not explicitly set. Split the words if the default value was used and
	// split_words was set to true
	switch newField.SecretType {
	case "map":
		newField.MapKeyName, ok, err = parseOptionalStructTagKey[string](fStructField, TagKeyName)
		if err != nil {
			return field{}, StructTagError{
				name: fStructField.Name,
				key:  TagKeyName,
				err:  err,
			}
		}
		if !ok {
			newField.MapKeyName = fStructField.Name
			if splitWordsEnabled {
				newField.MapKeyName = splitWords(newField.MapKeyName)
			}
		}
	default:
		if _, ok = fStructField.Tag.Lookup(TagKeyName); ok {
			return field{}, StructTagError{
				name: fStructField.Name,
				key:  TagKeyName,
				err:  ErrSecretTypeDoesNotSupportTagKey,
			}
		}
	}

	// Get the version value, setting it to the default, "default", if not explicitly
	// set. Split the words if the default value was used and split_words was set to true
	newField.SecretVersion, ok, err = parseOptionalStructTagKey[string](fStructField, TagVersion)
	if err != nil {
		return field{}, StructTagError{
			name: fStructField.Name,
			key:  TagVersion,
			err:  err,
		}
	}
	if !ok {
		newField.SecretVersion = DefaultVersion
	}

	return newField, nil
}

func splitWords(s string) string {
	return RegexMatchCapitals.ReplaceAllString(s, "${1}_${2}")
}
