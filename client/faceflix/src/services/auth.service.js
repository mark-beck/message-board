import axios from "axios";

const API_URL = "https://localhost/auth";


const login = async (email, password) => {
    const response = await axios
        .post(API_URL + '/signin', {
            email: email,
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

export const getCurrentUser = () => {
    return JSON.parse(localStorage.getItem("user"));
};

export default {
    login,
    logout,
    getCurrentUser,
};
