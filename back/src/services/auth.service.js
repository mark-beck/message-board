import axios from "axios";

const API_URL = "https://localhost/auth";


const login = async (username, password) => {
    const response = await axios
        .post(API_URL + '/signin', {
            name: username,
            password: password,
        });
    if (response.data.token) {
        console.log("token response" + response.data)
        localStorage.setItem("user", JSON.stringify(response.data));
    }
    return response.data;
};

const logout = () => {
    localStorage.removeItem("user");
};

const getCurrentUser = () => {
    return JSON.parse(localStorage.getItem("user"));
};

export default {
    login,
    logout,
    getCurrentUser,
};
