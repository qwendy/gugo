// convert post.md to post.html
package generator

type Post struct {
	SourcePath  string
	Destination string
}

func NewPost(sourcePath string) *Post {
	return &Post{
		SourcePath: sourcePath,
	}
}

func Convert() {

}
