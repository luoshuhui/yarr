package notion

// PageCreateRequest Notion创建页面请求
type PageCreateRequest struct {
	Parent     Parent     `json:"parent"`
	Properties Properties `json:"properties"`
	Children   []Block    `json:"children"`
}

type Parent struct {
	DatabaseID string `json:"database_id"`
}

type Properties struct {
	Title TitleProperty `json:"Name"` // Notion默认title属性名为"Name"
}

type TitleProperty struct {
	Title []RichText `json:"title"`
}

type RichText struct {
	Text Text `json:"text"`
}

type Text struct {
	Content string `json:"content"`
}

// Block Notion内容块
type Block struct {
	Object           string                  `json:"object"`
	Type             string                  `json:"type"`
	Paragraph        *ParagraphBlock         `json:"paragraph,omitempty"`
	Heading1         *HeadingBlock           `json:"heading_1,omitempty"`
	Heading2         *HeadingBlock           `json:"heading_2,omitempty"`
	Heading3         *HeadingBlock           `json:"heading_3,omitempty"`
	BulletedListItem *BulletedListItemBlock  `json:"bulleted_list_item,omitempty"`
	NumberedListItem *NumberedListItemBlock  `json:"numbered_list_item,omitempty"`
	Quote            *QuoteBlock             `json:"quote,omitempty"`
	Code             *CodeBlock              `json:"code,omitempty"`
}

type ParagraphBlock struct {
	RichText []RichText `json:"rich_text"`
}

type HeadingBlock struct {
	RichText []RichText `json:"rich_text"`
}

type BulletedListItemBlock struct {
	RichText []RichText `json:"rich_text"`
}

type NumberedListItemBlock struct {
	RichText []RichText `json:"rich_text"`
}

type QuoteBlock struct {
	RichText []RichText `json:"rich_text"`
}

type CodeBlock struct {
	RichText []RichText `json:"rich_text"`
	Language string     `json:"language"`
}

// PageCreateResponse Notion创建页面响应
type PageCreateResponse struct {
	Object string `json:"object"`
	ID     string `json:"id"`
	URL    string `json:"url"`
}
