import * as React from "react"
import {
  Show,
  SimpleShowLayout,
  TextField,
  EditButton,
  TopToolbar,
} from "react-admin"


export const CategoryShow = () =>
  <Show >
    <SimpleShowLayout>
      <TextField source="id" />
      <TextField source="title" />
    </SimpleShowLayout>
  </Show>

