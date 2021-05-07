//
//  AppDelegate.swift
//  dsk
//
//  Created by Marius Wilms on 05.05.21.
//

import Cocoa
import SwiftUI

@main
class AppDelegate: NSObject, NSApplicationDelegate {

    var window: NSWindow!

    var backend: Process!
    
    var ddt: String!

    func applicationDidFinishLaunching(_ aNotification: Notification) {
        // Create the SwiftUI view that provides the window contents.
        let contentView = ContentView()

        // Create the window and set the content view.
        window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 1000, height: 800),
            styleMask: [.titled, .closable, .miniaturizable, .resizable, .fullSizeContentView],
            backing: .buffered, defer: false)
        window.isReleasedWhenClosed = false
        window.center()
        window.setFrameAutosaveName("Main Window")
        window.contentView = NSHostingView(rootView: contentView)
        window.makeKeyAndOrderFront(nil)
        

        //runBackend()

    }
    
    func runBackend() {
        print("starting backend")

        if backend != nil && backend.isRunning {
            print("STOPPING backend")
            backend.interrupt()
            backend.waitUntilExit()
        }

        backend = Process()
        backend.executableURL = Bundle.main.url(forResource: "dsk-darwin-amd64", withExtension: "")

        backend.arguments = ["-port", "8080", ddt]
        try! backend.run()
    }

    func applicationWillTerminate(_ aNotification: Notification) {
        if backend != nil && backend.isRunning {
            backend.interrupt()
            backend.waitUntilExit()
        }
        // Insert code here to tear down your application
    }

    @IBAction func actionOpen(_ sender: NSMenuItem) {
        let dialog = NSOpenPanel();

        dialog.title                   = "Choose a design definitions tree";
        dialog.showsResizeIndicator    = true;
        dialog.showsHiddenFiles        = false;
        dialog.canChooseFiles = false;
        dialog.canChooseDirectories = true;

        if (dialog.runModal() ==  NSApplication.ModalResponse.OK) {
            let result = dialog.url

            if (result != nil) {
                let path: String = result!.path
                print(path)
                ddt = path
                runBackend()
                // path contains the directory path e.g
                // /Users/ourcodeworld/Desktop/folder
            }
        } else {
            // User clicked on "Cancel"
            return
        }
    
    }
    
}

