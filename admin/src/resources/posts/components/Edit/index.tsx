import {
  BooleanInput,
  Edit,
  RadioButtonGroupInput,
  SimpleForm,
  TextInput,
} from "react-admin"
import * as React from "react"

export const CategoryEdit = (props) => {
  return (
    <Edit {...props}>
      <SimpleForm redirect="list">
        <TextInput source="title" label="الاسم" />
        <TextInput source="description" label="الاسم" />
      </SimpleForm>
    </Edit>
  )
}
