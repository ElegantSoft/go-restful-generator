import * as React from "react"
import {
  Show,
  SimpleShowLayout,
  TextField,
  EditButton,
  TopToolbar,
} from "react-admin"

const PostShowActions = ({ basePath, data }) => (
  <TopToolbar>
    <EditButton basePath={basePath} record={data} />
  </TopToolbar>
)

export const CategoryShow = (props) => (
  <Show {...props} actions={<PostShowActions {...props} />}>
    <SimpleShowLayout>
      <TextField source="id" />
      <TextField source="name" />
    </SimpleShowLayout>
  </Show>
)
