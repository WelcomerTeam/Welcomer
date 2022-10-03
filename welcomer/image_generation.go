package welcomer

//go:generate go-enum -f=$GOFILE --marshal

// ENUM(left, center, right, topLeft, topCenter, topRight, bottomLeft, bottomCenter, bottomRight)
type ImageAlignment int32

// ENUM(default, vertical, card)
type ImageTheme int32

// ENUM(circular, rounded, squared, hexagonal)
type ImageProfileBorderType int32
