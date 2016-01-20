package gorma

import (
	"text/template"

	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen/codegen"
)

type (
	TypeConverterData struct {
		Type       *design.UserTypeDefinition
		UpperName  string
		LowerName  string
		Version    string
		VersionPkg string
	}

	// UserTypeTemplateData contains all the information used by the template to redner the
	// media types code.
	UserTypeTemplateData struct {
		ConvertTypes  map[string]TypeConverterData
		APIDefinition *design.APIDefinition
		UserType      *RelationalModelDefinition
		DefaultPkg    string
		AppPkg        string
	}
	// UserTypesWriter generate code for a goa application user types.
	// User types are data structures defined in the DSL with "Type".
	UserTypesWriter struct {
		*codegen.SourceFile
		UserTypeTmpl *template.Template
	}
)

// NewUserTypesWriter returns a contexts code writer.
// User types contain custom data structured defined in the DSL with "Type".
func NewUserTypesWriter(filename string) (*UserTypesWriter, error) {
	file, err := codegen.SourceFileFor(filename)
	if err != nil {
		return nil, err
	}
	return &UserTypesWriter{SourceFile: file}, nil
}

// Execute writes the code for the context types to the writer.
func (w *UserTypesWriter) Execute(data *UserTypeTemplateData) error {
	return w.ExecuteTemplate("types", userTypeT, nil, data)
}

// arrayAttribute returns the array element attribute definition.
func arrayAttribute(a *design.AttributeDefinition) *design.AttributeDefinition {
	return a.Type.(*design.Array).ElemType
}

const (
	// userTypeT generates the code for a user type.
	// template input: UserTypeTemplateData
	userTypeT = `// {{if .UserType.Description}}{{.UserType.Description}}
// {{if .UserType.ModeledType  }} // Stores {{.UserType.ModeledType.TypeName}}{{end}}{{else}}{{.UserType.Name }}
// {{if .UserType.ModeledType  }}Stores {{.UserType.ModeledType.TypeName}}{{end}}
type{{end}}
	{{.UserType.StructDefinition}}
{{ if ne .UserType.TableName "" }}
// TableName overrides the table name settings in gorm
func (m {{.UserType.Name}}) TableName() string {
	return "{{ .UserType.TableName}}"
}{{end}}
// {{.UserType.Name}}DB is the implementation of the storage interface for {{.UserType.Name}}
type {{.UserType.Name}}DB struct {
	Db gorm.DB
	{{ if .UserType.Cached }}cache *cache.Cache{{end}}
}
// New{{.UserType.Name}}DB creates a new storage type
func New{{.UserType.Name}}DB(db gorm.DB) *{{.UserType.Name}}DB {
	{{ if .UserType.Cached }}return &{{.UserType.Name}}DB{
		Db: db,
		cache: cache.New(5*time.Minute, 30*time.Second),
	}
	{{ else  }}return &{{.UserType.Name}}DB{Db: db}{{ end  }}
}
// DB returns  the underlying database
func (m *{{.UserType.Name}}DB) DB() interface{} {
	return &m.Db
}
{{ if .UserType.Roler }}
// GetRole returns the value of the role field and satisfies the Roler interface
func (m {{.UserType.Name}}) GetRole() string {
	return {{$f := .UserType.Fields.role}}{{if $f.Nullable}}*{{end}}m.Role
}
{{end}}

	 
// Storage Interface
type {{.UserType.Name}}Storage interface {
	DB() interface{}
	List(ctx context.Context{{ if .UserType.DynamicTableName}}, tableName string{{ end }}) []{{.UserType.Name}}
	One(ctx context.Context{{ if .UserType.DynamicTableName }}, tableName string{{ end }}, {{.UserType.PKAttributes}}) ({{.UserType.Name}}, error)
	Add(ctx context.Context{{ if .UserType.DynamicTableName }}, tableName string{{ end }}, {{.UserType.LowerName}} {{.UserType.Name}}) ({{.UserType.Name}}, error)
	Update(ctx context.Context{{ if .UserType.DynamicTableName }}, tableName string{{ end }}, {{.UserType.LowerName}} {{.UserType.Name}}) (error)
	Delete(ctx context.Context{{ if .UserType.DynamicTableName }}, tableName string{{ end }}, {{ .UserType.PKAttributes}}) (error) 	{{$typename:= .UserType.Name}}{{$dtn:=.UserType.DynamicTableName}}{{ range $idx, $bt := .UserType.BelongsTo}}
	ListBy{{$bt.Name}}(ctx context.Context{{ if $dtn}}, tableName string{{ end }},{{$bt.LowerName}}_id int) []{{$typename}}
	OneBy{{$bt.Name}}(ctx context.Context{{ if $dtn}}, tableName string{{ end }}, {{$bt.LowerName}}_id, id int) ({{$typename}}, error){{end}}
	{{range $i, $m2m := .UserType.ManyToMany}}
	List{{$m2m.RightNamePlural}}(context.Context, int) []{{$m2m.RightName}}.{{$m2m.RightName}}
	Add{{$m2m.RightNamePlural}}(context.Context, int, int) (error)
	Delete{{$m2m.RightNamePlural}}(context.Context, int, int) error
	{{end}}
}

// CRUD Functions
// One returns a single record by ID
func (m *{{$typename}}DB) One(ctx context.Context{{ if .UserType.DynamicTableName }}, tableName string{{ end }}, {{.UserType.PKAttributes}}) ({{$typename}}, error) {
	{{ if .UserType.Cached }}//first attempt to retrieve from cache
	o,found := m.cache.Get(strconv.Itoa(id))
	if found {
		return o.({{$typename}}), nil
	}
	// fallback to database if not found{{ end }}
	var obj {{$typename}}{{ $l := len $.UserType.PrimaryKeys }}{{ if eq $l 1 }}
	err := m.Db{{ if .UserType.DynamicTableName }}.Table(tableName){{ end }}.Find(&obj, id).Error{{ else  }}err := m.Db{{ if .UserType.DynamicTableName }}.Table(tableName){{ end }}.Find(&obj).Where("{{.UserType.PKWhere}}", {{.UserType.PKWhereFields }} id).Error{{ end }}
	{{ if .UserType.Cached }} go m.cache.Set(strconv.Itoa(id), obj, cache.DefaultExpiration) {{ end }}
	return obj, err
}
// Add creates a new record
func (m *{{$typename}}DB) Add(ctx context.Context{{ if .UserType.DynamicTableName }}, tableName string{{ end }}, model {{$typename}}) ({{$typename}}, error) {
	err := m.Db{{ if .UserType.DynamicTableName }}.Table(tableName){{ end }}.Create(&model).Error{{ if .UserType.Cached }} 
	go m.cache.Set(strconv.Itoa(model.ID), model, cache.DefaultExpiration) {{ end }}
	return model, err
}
// Update modifies a single record
func (m *{{$typename}}DB) Update(ctx context.Context{{ if .UserType.DynamicTableName }}, tableName string{{ end }}, model {{$typename}}) error {
	obj, err := m.One(ctx{{ if .UserType.DynamicTableName }}, tableName{{ end }}, {{.UserType.PKUpdateFields}})
	if err != nil {
		return  err
	}
	err = m.Db{{ if .UserType.DynamicTableName }}.Table(tableName){{ end }}.Model(&obj).Updates(model).Error
	{{ if .UserType.Cached }}go func(){
	obj, err := m.One(ctx, model.ID)
	if err == nil {
		m.cache.Set(strconv.Itoa(model.ID), obj, cache.DefaultExpiration)
	}
	}()
	{{ end }}
	return err
}
// Delete removes a single record
func (m *{{$typename}}DB) Delete(ctx context.Context{{ if .UserType.DynamicTableName }}, tableName string{{ end }}, {{.UserType.PKAttributes}})  error {
	var obj {{$typename}}{{ $l := len .UserType.PrimaryKeys }}{{ if eq $l 1 }}
	err := m.Db{{ if .UserType.DynamicTableName }}.Table(tableName){{ end }}.Delete(&obj, id).Error
	{{ else  }}err := m.Db{{ if .UserType.DynamicTableName }}.Table(tableName){{ end }}.Delete(&obj).Where("{{.UserType.PKWhere}}", {{.UserType.PKWhereFields}}).Error
	{{ end }}
	if err != nil {
		return  err
	}
	{{ if .UserType.Cached }} go m.cache.Delete(strconv.Itoa(id)) {{ end }}
	return  nil
} 
{{$ut := .UserType}}{{$typename := .UserType.Name}}{{ range $idx, $bt := .UserType.BelongsTo}}
// Belongs To Relationships
// {{$typename}}FilterBy{{$bt.Name}} is a gorm filter for a Belongs To relationship
func {{$typename}}FilterBy{{$bt.Name}}(parentid int, originaldb *gorm.DB) func(db *gorm.DB) *gorm.DB {
	if parentid > 0 {
		return func(db *gorm.DB) *gorm.DB {
			return db.Where("{{$bt.LowerName}}_id", parentid)
		}
	} else {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
}
// ListBy{{$bt.Name}} returns an array of associated {{$bt.Name}} models
func (m *{{$typename}}DB) ListBy{{$bt.Name}}(ctx context.Context{{ if $ut.DynamicTableName }}, tableName string{{ end }}, parentid int) []{{$typename}} {
	var objs []{{$typename}}
	m.Db{{ if $ut.DynamicTableName }}.Table(tableName){{ end }}.Scopes({{$typename}}FilterBy{{$bt.Name}}(parentid, &m.Db)).Find(&objs)
	return objs
}
// OneBy{{$bt.Name}} returns a single associated {{$bt.Name}} model
func (m *{{$typename}}DB) OneBy{{$bt.Name}}(ctx context.Context{{ if $ut.DynamicTableName }}, tableName string{{ end }}, parentid, {{ $ut.PKAttributes}}) ({{$typename}}, error) {
	{{ if $ut.Cached }}//first attempt to retrieve from cache
	o,found := m.cache.Get(strconv.Itoa(id))
	if found {
		return o.({{$typename}}), nil
	}
	// fallback to database if not found{{ end }}
	var obj {{$typename}}
	err := m.Db{{ if $ut.DynamicTableName }}.Table(tableName){{ end }}.Scopes({{$typename}}FilterBy{{$bt.Name}}(parentid, &m.Db)).Find(&obj, id).Error
	{{ if $ut.Cached }} go m.cache.Set(strconv.Itoa(id), obj, cache.DefaultExpiration) {{ end }}
	return obj, err
}
{{end}} 

{{$ut := .UserType }}{{$typeName := .UserType.Name}}{{ range $idx, $bt := .UserType.ManyToMany}}
// Many To Many Relationships
// Delete{{goify $bt.RightName true}} removes a {{$bt.RightName}}/{{$bt.LeftName}} entry from the join table
func (m *{{$typeName}}DB) Delete{{goify $bt.RightName true}}(ctx context.Context{{ if $ut.DynamicTableName }}, tableName string{{ end }}, {{$ut.Lower}}ID,  {{$bt.LowerRightName}}ID int)  error {
	var obj {{$typeName}}
	obj.ID = {{$ut.LowerName}}ID
	var assoc {{$bt.LowerRightName}}.{{$bt.RightName}}
	var err error
	assoc.ID = {{$bt.LowerRightName}}ID
	if err != nil {
		return err
	}
	err = m.Db{{ if $ut.DynamicTableName }}.Table(tableName){{ end }}.Model(&obj).Association("{{$bt.RightNamePlural}}").Delete(assoc).Error
	if err != nil {
		return  err
	}
	return  nil
}  
// Add{{goify $bt.RightName true}} creates a new {{$bt.RightName}}/{{$bt.LeftName}} entry in the join table
func (m *{{$typeName}}DB) Add{{goify $bt.RightName true}}(ctx context.Context{{ if $ut.DynamicTableName }}, tableName string{{ end }}, {{$ut.LowerName}}ID, {{$bt.LowerRightName}}ID int) error {
	var {{$ut.LowerName}} {{$typeName}}
	{{$ut.LowerName}}.ID = {{$ut.LowerName}}ID
	var assoc {{$bt.LowerRightName}}.{{$bt.RightName}}
	assoc.ID = {{$bt.LowerRightName}}ID
	err := m.Db{{ if $ut.DynamicTableName }}.Table(tableName){{ end }}.Model(&{{$ut.Lower}}).Association("{{$bt.RightNamePlural}}").Append(assoc).Error
	if err != nil {
		return  err
	}
	return  nil
}
// List{{goify $bt.RightName true}} returns a list of the {{$bt.RightName}} models related to this {{$bt.LeftName}}
func (m *{{$typeName}}DB) List{{goify $bt.RightName true}}(ctx context.Context{{ if $ut.DynamicTableName }}, tableName string{{ end }}, {{$ut.Lower}}ID int)  []{{$bt.LowerRightName}}.{{$bt.RightName}} {
	var list []{{$bt.LowerRightName}}.{{$bt.RightName}}
	var obj {{$typeName}}
	obj.ID = {{$ut.LowerName}}ID
	m.Db{{ if $ut.DynamicTableName }}.Table(tableName){{ end }}.Model(&obj).Association("{{$bt.RightNamePlural}}").Find(&list)
	return  list
}
{{end}}
{{ range $idx, $bt := .UserType.BelongsTo}}
// Filter{{$typename}}By{{$bt.Name}} iterates a list and returns only those with the foreign key provided
func Filter{{$typename}}By{{$bt.Name}}(parent *int, list []{{$typename}}) []{{$typename}} {
	var filtered []{{$typename}}
	for _,o := range list {
		if o.{{$bt.Name}}ID == int(*parent) {
			filtered = append(filtered,o)
		}
	}
	return filtered
}
{{end}}

{{ if .UserType.ModeledType }}{{$ut := .UserType}}
// Useful conversion functions
func (m *{{$typeName}}DB) To{{.UserType.ModeledType.TypeName}}() {{.AppPkg}}.{{.UserType.ModeledType.TypeName}} {
	payload := {{.AppPkg}}.{{.UserType.ModeledType.TypeName}}{}
	{{ range $fname, $field := $ut.RelationalFields }}{{$obj  := $ut.ModeledType.ToObject}}{{range $key, $def := $obj}} {{if eq $field.LowerName $key}}payload.{{title $key}} = m.{{$fname}}
	{{end}}{{end}}{{ end }}	return payload
}
{{end}}

{{$ut := .UserType}}{{ range $key, $tcd := .ConvertTypes }}
// Convert from	{{if eq $tcd.Version ""}}default version{{else}}Version {{$tcd.Version}}{{end}} {{$tcd.UpperName}} to {{$typeName}}
func {{$typeName}}From{{$tcd.Version}}{{$tcd.UpperName}}(t {{if eq $tcd.Version ""}}{{$tcd.Version}}.{{end}}{{$tcd.UpperName}}) {{$typeName}} {
	{{$ut.LowerName}} := {{$ut.Name}}{}
	{{ range $fname, $field := $ut.RelationalFields }}{{$obj  := $tcd.Type.ToObject}}{{range $key, $def := $obj}} {{if eq $field.LowerName $key}}{{$ut.LowerName}}.{{title $key}} = t.{{$fname}}
	{{end}}{{end}}{{ end }}	
	return {{$ut.LowerName}}
}
{{end}}
`
)
