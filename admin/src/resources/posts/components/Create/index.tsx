import React from "react"
import { Create, SimpleForm, TextInput } from "react-admin"

export const CategoryCreate = (props) => {
  return (
    <Create {...props}>
      <SimpleForm redirect="list">
        <TextInput source="name" label="الاسم" />
      </SimpleForm>
    </Create>
  )
}
