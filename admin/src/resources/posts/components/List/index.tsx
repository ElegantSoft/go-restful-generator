import React, { FC } from "react"
import {
  Datagrid,
  List,
  TextField,
  Pagination,
  Filter,
  TextInput,
  SingleFieldList,
  NumberInput,
  ChipField,
  BooleanField,
  ArrayField,
  DatagridProps,
} from "react-admin"
import { MobileGrid } from "../MobileGrid"

const PostPagination = () => (
  <Pagination rowsPerPageOptions={[25, 50, 100]} />
)


// const BlacklistFilter = (props) => (
//   <Filter {...props}>
//     <NumberInput label="المعرف" source="id||eq" alwaysOn />
//     <TextInput source="name" alwaysOn label="الاسم" />
//   </Filter>
// )

const filters = [
  <NumberInput label="المعرف" source="id||eq" alwaysOn />,
  <TextInput source="title" label="الاسم" alwaysOn />,
]

export const CategoryList = () => {

  return (
    <List
      pagination={<PostPagination />}
      filters={filters}
      hasCreate
      perPage={25}
    >
      <Datagrid rowClick="edit" optimized>
        <TextField source="id" label="المعرف" />
        <TextField source="title" label="الاسم" />
        <TextField source="category.name" label="اسم القسم" />
      </Datagrid>
    </List>
  )
}
