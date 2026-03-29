const dbHost = process.env.DB_HOST;
const dbPort = process.env['DB_PORT'];
const apiKey = process.env.API_KEY;

const config = {
  database: `postgres://${dbHost}:${dbPort}`,
  apiKey: apiKey
};

console.log(config);
