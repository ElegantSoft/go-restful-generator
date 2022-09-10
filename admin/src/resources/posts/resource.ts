import { CategoryList } from "./components/List"
import { CategoryCreate } from "./components/Create"
import { CategoryEdit } from "./components/Edit"
import { CategoryShow } from "./components/Show"

export const postsResource = {
  list: CategoryList,
  show: CategoryShow,
  edit: CategoryEdit,
  create: CategoryCreate,
}
