import { useState } from 'react'
import { Admin, fetchUtils, Resource } from 'react-admin'
import { apiProvider } from './config/api.provider'
import { API_URL } from './config/constants'
import { postsResource } from './resources/posts/resource'

const httpClient = (url: string, options: any = {}) => {
  if (!options.headers) {
    options.headers = new Headers({ Accept: "application/json" })
  }
  const token = localStorage.getItem("token")
  options.headers.set("Authorization", `Bearer ${token}`)
  return fetchUtils.fetchJson(url, options)
}

const dataProvider = apiProvider(API_URL, httpClient)

function App() {
  const [count, setCount] = useState(0)


  return (
    <Admin dataProvider={dataProvider}>
      <Resource name="posts" {...postsResource} />
    </Admin>
  )
}

export default App
