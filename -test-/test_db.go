package main

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

/**************************************/
// {
//     "field": [
//         {
//             "order": "该阶段下字段的顺序编号数字表示,数字越小,优先级越高1",
//             "field_id": "1",
//             "status": "字段状态，1有效 2无效（删除）",
//             "text": {
//                 "title": "文本标题",
//                 "one_or_many": "单行或者多行文本 one or many",
//                 "subhead_open": "文本副标题是否开启，true or false",
//                 "subhead": "文本副标题,对主标题的补充说明最长100字符",
//                 "helptext_open": "帮助说明是否开启，true or false",
//                 "helptext": "帮助说明",
//                 "data": "字段数据,文本内容",
//                 "necessary": "是否必填,true or false",
//                 "other_phase_edit": "是否允许其它阶段编辑,true or false",
//                 "created_time": "2017-05-10 15:09:54",
//                 "updated_time": "2017-05-10 15:09:54"
//             }
//         },
//         {
//             "order": "该阶段下字段的顺序编号数字表示,数字越小,优先级越高2",
//             "field_id": "2",
//             "status": "字段状态，1有效 2无效（删除）",
//             "text": {
//                 "title": "文本标题",
//                 "one_or_many": "单行或者多行文本 one or many",
//                 "subhead_open": "文本副标题是否开启，true or false",
//                 "subhead": "文本副标题,对主标题的补充说明最长100字符",
//                 "helptext_open": "帮助说明是否开启，true or false",
//                 "helptext": "帮助说明",
//                 "data": "字段数据,文本内容",
//                 "necessary": "是否必填,true or false",
//                 "other_phase_edit": "是否允许其它阶段编辑,true or false",
//                 "created_time": "2017-05-10 15:09:54",
//                 "updated_time": "2017-05-10 15:09:54"
//             }
//         },
//         {
//             "order": "该阶段下字段的顺序编号数字表示,数字越小,优先级越高3",
//             "field_id": "3",
//             "status": "字段状态，1有效 2无效（删除）",
//             "describe": {
//                 "data": {
//                     "context": "描述内容"
//                 },
//                 "created_time": "2017-05-10 15:09:54",
//                 "updated_time": "2017-05-10 15:09:54"
//             }
//         }
//     ]
// }
/**************************************/
type Fields struct {
	Id   int
	List []Field
}
type Field struct {
	Order    int
	FieldId  int
	Describe Describe
	Text     Text
}
type Describe struct {
	Data        Data
	CreatedTime string
	UpdatedTime string
}
type Text struct {
	Title       string
	Data        string
	CreatedTime string
	UpdatedTime string
}
type Data struct {
	Context string
}

func main() {
	session, err := mgo.Dial("")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("test").C("kong")

	/************插入数据*************/
	field1 := Field{
		Order:    1,
		FieldId:  2,
		Describe: describe1,
	}
	field2 := Field{
		Order:   2,
		FieldId: 3,
		Text:    text1,
	}
	field3 := Field{
		Order:   3,
		FieldId: 4,
		Text:    text2,
	}
	describe1 := Describe{
		Data:        data1,
		CreatedTime: "2000",
		UpdatedTime: "2000",
	}
	text1 := Text{
		Title:       "标题",
		Data:        "string1111",
		CreatedTime: "2000",
		UpdatedTime: "2000",
	}
	text2 := Text{
		Title:       "标题",
		Data:        "string222222",
		CreatedTime: "9999",
		UpdatedTime: "9999",
	}

	fields := Fields{
		Id:   100,
		List: []Field{field1, field2, field3},
	}
	/************插入数据*************/

	err = c.Insert(*fields)
	if err != nil {
		panic(err)
	}

	result := Fields{}
	err = c.Find(bson.M{"name": "Ale"}).One(&result)
	if err != nil {
		panic(err)
	}

	fmt.Println("Phone:", result.Phone)
}
