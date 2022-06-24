import axios from "axios";
import authHeader from "./auth-header";

const API_URL = "https://localhost/admin";
const C_URL = "https://localhost/content";


const getUsers = () => {
    return axios.get(API_URL + "/list_users", { headers: authHeader() }).then((users) => {
        if (typeof (users) == Object) {
            users = Array(1).fill(users)
        }
        return users;
    })
    
};

const createUser = (name, email, password, user, moderator, admin) => {
    let roles = Array(0);
    if (user) {
        roles.push("User")
    }
    if (moderator) {
        roles.push("Moderator")
    }
    if (admin) {
        roles.push("Admin")
    }

    const create_user_json = {
        name: name,
        email: email,
        password: password,
        roles: roles,
    }
    return axios.post(API_URL + "/create_user", create_user_json, { headers: authHeader() });
    
};

const deleteUser = (name) => {
    console.log("deleteUser called");
    return axios.delete(API_URL + "/delete_user" + name, { headers: authHeader() });
}

export async function getPublicContent() {
    console.log("getting content");
    
    let posts = await axios.get(C_URL + "/all", { headers: authHeader() });

    return posts.data

    
};

export async function postContent(text) {
    let author = JSON.parse(localStorage.getItem("user")).user.name
    console.log("User:: " + author)
    const post = {
        author: author,
        text: text,
        date: ""
    }

    return axios.post(C_URL + "/add", post, { headers: authHeader() });
}

export async function getFilteredContent(filter) {
    console.log("getting filtered content");

    let posts = await axios.post(C_URL + "/filter", filter, { headers: authHeader() });

    console.log(posts);
    if (posts.data == null) {
        return new Array(0);
    }
    return posts.data;

}

// a function that converts ISO 8601 dates to a readable format
export function formatDate(date) {
    var d = new Date(date);

    var diffMs = (new Date() - d);

    var diffMins = Math.round(diffMs/ 60000);

    if (diffMins < 60) {
        return diffMins + " minutes ago";
    }

    if (diffMins < 60 * 24) {
        return Math.round(diffMins / 60) + " hours ago";
    }
    return `${d.getHours()}:${d.getMinutes()} ${d.getDate()}.${d.getMonth()}.${d.getFullYear()}`;
}


export async function delete_post(id) {
    return await axios.delete(C_URL + "/delete/" + id, { headers: authHeader() });
}


export default {
    getUsers,
    createUser,
    deleteUser,
    getPublicContent,
    postContent,
    getFilteredContent,
    formatDate,
};
