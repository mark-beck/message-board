import axios from "axios";
import authHeader from "./auth-header";

const API_URL = "https://localhost/admin";


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
    return axios.delete(API_URL + "/delete_user/" + name, { headers: authHeader() });
}

const getPublicContent = () => {
    return "Public Content";
};



export default {
    getUsers,
    createUser,
    deleteUser,
    getPublicContent,
};
