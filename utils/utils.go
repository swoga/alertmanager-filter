package utils

// StringArray is an array of strings
type StringArray []string

// Contains checks if the supplied string is contained in the array
func (list StringArray) Contains(search string) bool {
	for _, value := range list {
		if value == search {
			return true
		}
	}
	return false
}
