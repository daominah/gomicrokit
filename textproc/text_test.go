package textproc

import (
	"encoding/json"
	"math"
	"testing"
)

func TestTextToNGrams(t *testing.T) {
	text := `Có thánh này, chắc chắn "Sẻ đệ" (NDB 2.0) sẽ thêm sức mạnh để đả bại Sơ Luyến.
Trực tiếp ngay bây giờ trên http://www.gametv1.vn. ______ Ahihi`
	words := TextToWords(text)
	jbs, err := json.Marshal(words)
	if err != nil {
		t.Error(err)
	}
	if string(jbs) != `["Có","thánh","này","chắc","chắn","Sẻ","đệ","NDB","2.0","sẽ","thêm","sức","mạnh","để","đả","bại","Sơ","Luyến","Trực","tiếp","ngay","bây","giờ","trên","http://www.gametv1.vn","Ahihi"]` {
		t.Error(string(jbs))
	}

	nGrams := TextToNGrams(text, 3)
	jbs, err = json.Marshal(nGrams)
	if err != nil {
		t.Error(err)
	}
	if string(jbs) != `{"2.0 sẽ thêm":1,"bây giờ trên":1,"bại sơ luyến":1,"chắc chắn sẻ":1,"chắn sẻ đệ":1,"có thánh này":1,"giờ trên http://www.gametv1.vn":1,"luyến trực tiếp":1,"mạnh để đả":1,"ndb 2.0 sẽ":1,"ngay bây giờ":1,"này chắc chắn":1,"sơ luyến trực":1,"sẻ đệ ndb":1,"sẽ thêm sức":1,"sức mạnh để":1,"thánh này chắc":1,"thêm sức mạnh":1,"tiếp ngay bây":1,"trên http://www.gametv1.vn ahihi":1,"trực tiếp ngay":1,"đả bại sơ":1,"để đả bại":1,"đệ ndb 2.0":1}` {
		t.Error(string(jbs))
	}
}

func TestHashTextToInt64(t *testing.T) {
	nWords := 100000
	words := make(map[string]bool)
	hashes := make(map[int64]bool)
	for i := 0; i < nWords; i++ {
		word := GenRandomWord(1, 4)
		words[word] = true
		hashes[HashTextToInt(word)] = true
		hashes[HashTextToInt(word)] = true
	}
	if math.Abs(float64(len(words)-len(hashes))) > 10 {
		t.Error()
	}
}

func TestTextNormalize(t *testing.T) {
	in := `VE bị phạt 100 triệu đồng do không công bố thông tin đúng quy định
Ngày 30/09, Thanh tra Ủy ban Chứng khoán Nhà nước (UBCKNN) đã quyết định xử phạt
vi phạm hành chính trong lĩnh vực chứng khoán và thị trường chứng khoán đối với 
Tổng Công ty Tư vấn thiết kế dầu khí - CTCP (HNX:PVE). 
Cụ thể, Công ty này đã không công bố thông tin tài liệu.`
	out := NormalizeText(in)
	if out != `VE bị phạt 100 triệu đồng do không công bố thông tin đúng quy định
Ngày 30/09, Thanh tra Ủy ban Chứng khoán Nhà nước (UBCKNN) đã quyết định xử phạt
vi phạm hành chính trong lĩnh vực chứng khoán và thị trường chứng khoán đối với 
Tổng Công ty Tư vấn thiết kế dầu khí - CTCP (HNX:PVE). 
Cụ thể, Công ty này đã không công bố thông tin tài liệu.` {
		t.Error(out)
	}
}
