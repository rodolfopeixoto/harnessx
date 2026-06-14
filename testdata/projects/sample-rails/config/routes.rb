Rails.application.routes.draw do
  root 'home#index'
  get '/about', to: 'pages#about'
  resources :products
  post '/api/orders', to: 'orders#create'
end
