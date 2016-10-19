package example

import _base "github.com/SphereSoftware/go-annotations/registry"
import a1 "github.com/SphereSoftware/go-annotations/example/test2"
import a2 "github.com/SphereSoftware/go-annotations/example/test"

func init() {
    _base.Map("github.com/SphereSoftware/go-annotations/example.Sample",
        _base.Annotations {
            Self: []interface{} {
                a2.Entity{
                    "",
                    nil,
                },
            },
            Fields: map[string][]interface{} {
            },
            Methods: map[string][]interface{} {
                "addSomething": []interface{} {
                    a2.Book{
                        "",
                        1.0,
                        nil,
                    },
},
}})
    _base.Map("github.com/SphereSoftware/go-annotations/example.JustAFunc",
        _base.Annotations {
            Self: []interface{} {
                a2.Book{
                    "",
                    1.0,
                    nil,
                },
            },
            Fields: map[string][]interface{} {
            },
            Methods: map[string][]interface{} {
}})
    _base.Map("github.com/SphereSoftware/go-annotations/example.TestAnotherFile",
        _base.Annotations {
            Self: []interface{} {
                a2.Entity{
                    "",
                    nil,
                },
            },
            Fields: map[string][]interface{} {
            },
            Methods: map[string][]interface{} {
}})
    _base.Map("github.com/SphereSoftware/go-annotations/example.Test",
        _base.Annotations {
            Self: []interface{} {
                a2.Entity{
                    "test",
                    []a2.Book{
                        a2.Book{
                            "book",
                            1.0,
                            &a1.Person{
                                "Mr.X",
                            },
                        },
                    },
                },
            },
            Fields: map[string][]interface{} {
                "methodOfTest": []interface{} {
                    a2.Entity{
                        "",
                        nil,
                    },
                },
            },
            Methods: map[string][]interface{} {
}})
    _base.Map("github.com/SphereSoftware/go-annotations/example.Test2",
        _base.Annotations {
            Self: []interface{} {
                a2.Entity{
                    "",
                    nil,
                },
            },
            Fields: map[string][]interface{} {
                "Name": []interface{} {
                    a2.Book{
                        "",
                        1.0,
                        nil,
                    },
                },
            },
            Methods: map[string][]interface{} {
}})

}
