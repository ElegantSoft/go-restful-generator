import {
  RequestQueryBuilder,
  QueryFilterArr,
  CondOperator,
} from "@nestjsx/crud-request";
import {
  CREATE,
  DELETE,
  DELETE_MANY,
  fetchUtils,
  GET_LIST,
  GET_MANY,
  GET_MANY_REFERENCE,
  GET_ONE,
  UPDATE,
  UPDATE_MANY,
} from "react-admin";

export const apiProvider = (apiUrl, httpClient = fetchUtils.fetchJson) => {
  const composeFilter = (paramsFilter) => {
    if (
      paramsFilter === "" ||
      (typeof paramsFilter.q !== "undefined" && paramsFilter.q === "")
    ) {
      paramsFilter = {};
    }

    const flatFilter = fetchUtils.flattenObject(paramsFilter);
    const filter = Object.keys(flatFilter).map((key) => {
      const splitKey = key.split("||");
      const ops = splitKey[1] ? splitKey[1] : "cont";
      let field = splitKey[0];

      if (field.indexOf("_") === 0 && field.indexOf(".") > -1) {
        field = field.split(/\.(.+)/)[1];
      }
      return { field, operator: ops, value: flatFilter[key] };
    });
    console.log("filter -> ", filter);
    return filter;
  };

  const convertDataRequestToHTTP = (type, resource, params) => {
    let url = "";
    const options: any = {};
    switch (type) {
      case GET_LIST: {
        const { page, perPage } = params.pagination;

        let query = RequestQueryBuilder.create({
          filter: composeFilter(params.filter) as QueryFilterArr,
        })
          .setLimit(perPage)
          .setPage(page)
          .sortBy(params.sort)
          .setOffset((page - 1) * perPage)
          .query();
        const regex = /filter%5B\d%5D/gi;
        console.log(query.replace(regex, "filter"));
        query = query.replaceAll(regex, "filter");
        if (resource === "moderators") {
          url = `${apiUrl}/users/moderators?${query}`;
        } else {
          url = `${apiUrl}/${resource}?${query}`;
        }

        break;
      }
      case GET_ONE: {
        if (resource === "moderators") {
          url = `${apiUrl}/users/${params.id}`;
        } else {
          url = `${apiUrl}/${resource}/${params.id}`;
        }
        break;
      }
      case GET_MANY: {
        const query = RequestQueryBuilder.create()
          .setFilter({
            field: "id",
            operator: CondOperator.IN,
            value: `${params.ids}`,
          })
          .query();

        if (resource === "moderators") {
          url = `${apiUrl}/users?${query}`;
        } else {
          url = `${apiUrl}/${resource}?${query}`;
        }
        break;
      }
      case GET_MANY_REFERENCE: {
        const { page, perPage } = params.pagination;
        const filter = composeFilter(params.filter) as QueryFilterArr;

        filter.push({
          field: params.target,
          operator: CondOperator.EQUALS,
          value: params.id,
        });

        const query = RequestQueryBuilder.create({
          filter,
        })
          .sortBy(params.sort)
          .setLimit(perPage)
          .setOffset((page - 1) * perPage)
          .query();

        if (resource === "moderators") {
          url = `${apiUrl}/users?${query}`;
        } else {
          url = `${apiUrl}/${resource}?${query}`;
        }
        break;
      }
      case UPDATE: {
        if (resource === "moderators") {
          url = `${apiUrl}/users/${params.id}`;
        } else {
          url = `${apiUrl}/${resource}/${params.id}`;
        }
        options.method = "PATCH";
        options.body = JSON.stringify(params.data);
        break;
      }
      case CREATE: {
        if (resource === "tickets") {
          const formData = new FormData();
          for (const param in params.data) {
            // when using multiple files
            if (param === "files") {
              params.data[param].forEach((file) => {
                formData.append("files", file.rawFile);
              });
              continue;
            }

            formData.append(param, params.data[param]);
          }
          url = `${apiUrl}/${resource}`;
          options.method = "POST";
          options.body = formData;
          break;
        } else if (resource === "moderators") {
          url = `${apiUrl}/users/admin-create`;
        } else {
          url = `${apiUrl}/${resource}`;
        }
        options.method = "POST";
        options.body = JSON.stringify(params.data);
        break;
      }
      case DELETE: {
        if (resource === "moderators") {
          url = `${apiUrl}/users/${params.id}`;
        } else {
          url = `${apiUrl}/${resource}/${params.id}`;
        }
        options.method = "DELETE";
        break;
      }
      default:
        throw new Error(`Unsupported fetch action type ${type}`);
    }
    return { url, options };
  };

  const convertHTTPResponse = (response, type, resource, params) => {
    const { headers, json } = response;
    switch (type) {
      case GET_LIST:
      case GET_MANY_REFERENCE:
        return {
          data: json.data,
          total: json.total,
        };
      case CREATE:
        return { data: { ...params.data, id: json.id } };
      default:
        return { data: json };
    }
  };

  return (type, resource, params) => {
    if (type === UPDATE_MANY) {
      return Promise.all(
        params.ids.map((id) =>
          httpClient(`${apiUrl}/${resource}/${id}`, {
            method: "PUT",
            body: JSON.stringify(params.data),
          })
        )
      ).then((responses) => ({
        data: responses.map((response) => response.json),
      }));
    }
    if (type === DELETE_MANY) {
      return Promise.all(
        params.ids.map((id) =>
          httpClient(`${apiUrl}/${resource}/${id}`, {
            method: "DELETE",
          })
        )
      ).then((responses) => ({
        data: responses.map((response) => response.json),
      }));
    }

    const { url, options } = convertDataRequestToHTTP(type, resource, params);
    return httpClient(url, options).then((response) =>
      convertHTTPResponse(response, type, resource, params)
    );
  };
};
