//
//  ContentView.swift
//  dsk
//
//  Created by Marius Wilms on 05.05.21.
//

import SwiftUI
import WebKit

// https://stackoverflow.com/questions/60945972/why-does-my-wkwebview-not-show-up-in-a-swiftui-view
struct ContentView: View {

    var body: some View {
        GeometryReader { g in
            ScrollView {
                WebView()
                .frame(height: g.size.height)
            }.frame(height: g.size.height)
        }
    }
}


struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}

// https://gist.github.com/joshbetz/2ff5922203240d4685d5bdb5ada79105

struct WebView: NSViewRepresentable {
    var request: URLRequest {
        get{
            let url: URL = URL(string: "http://127.0.0.1:8080")!
            let request: URLRequest = URLRequest(url: url)
            return request
        }
    }
    let view: WKWebView = WKWebView()
    
    func makeNSView(context: Context) -> WKWebView {
        view.load(request)
        return view
    }

    func updateNSView(_ view: WKWebView, context: Context) {
        view.load(request)
    }
}

struct WebView_Previews : PreviewProvider {
    static var previews: some View {
        WebView()
    }
}

